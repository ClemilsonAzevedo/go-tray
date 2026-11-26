package news

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maximumCachedResponses = 64

type cachedAPIResponse struct {
	response             Response
	entityTag            string
	receivedAt           time.Time
	maxAge               time.Duration
	staleWhileRevalidate time.Duration
	noCache              bool
	lastUsed             time.Time
}

func (entry cachedAPIResponse) isFresh(now time.Time) bool {
	return !entry.noCache && now.Sub(entry.receivedAt) < entry.maxAge
}

func (entry cachedAPIResponse) canServeStale(now time.Time) bool {
	if entry.noCache || entry.staleWhileRevalidate <= 0 {
		return false
	}

	age := now.Sub(entry.receivedAt)
	return age >= entry.maxAge && age-entry.maxAge < entry.staleWhileRevalidate
}

type httpResponseCache struct {
	mu            sync.Mutex
	entries       map[string]cachedAPIResponse
	revalidations map[string]chan struct{}
}

func (cache *httpResponseCache) get(key string) (cachedAPIResponse, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	entry, ok := cache.entries[key]
	if ok {
		entry.lastUsed = time.Now()
		cache.entries[key] = entry
	}
	return entry, ok
}

func (cache *httpResponseCache) store(key string, entry cachedAPIResponse) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.entries == nil {
		cache.entries = make(map[string]cachedAPIResponse)
	}
	if _, exists := cache.entries[key]; !exists && len(cache.entries) >= maximumCachedResponses {
		oldestKey := ""
		var oldest time.Time
		for candidate, cached := range cache.entries {
			if oldestKey == "" || cached.lastUsed.Before(oldest) {
				oldestKey = candidate
				oldest = cached.lastUsed
			}
		}
		delete(cache.entries, oldestKey)
	}
	entry.lastUsed = time.Now()
	cache.entries[key] = entry
}

func (cache *httpResponseCache) delete(key string) {
	cache.mu.Lock()
	delete(cache.entries, key)
	cache.mu.Unlock()
}

func (cache *httpResponseCache) beginRevalidation(key string) (chan struct{}, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.revalidations == nil {
		cache.revalidations = make(map[string]chan struct{})
	}
	if done, exists := cache.revalidations[key]; exists {
		return done, false
	}

	done := make(chan struct{})
	cache.revalidations[key] = done
	return done, true
}

func (cache *httpResponseCache) finishRevalidation(key string) {
	cache.mu.Lock()
	if done, exists := cache.revalidations[key]; exists {
		delete(cache.revalidations, key)
		close(done)
	}
	cache.mu.Unlock()
}

type cacheDirectives struct {
	present                 bool
	hasMaxAge               bool
	maxAge                  time.Duration
	hasStaleWhileRevalidate bool
	staleWhileRevalidate    time.Duration
	noCache                 bool
	noStore                 bool
}

func parseCacheDirectives(values []string) cacheDirectives {
	directives := cacheDirectives{present: len(values) > 0}
	for _, value := range values {
		for _, part := range splitCacheControl(value) {
			name, argument, hasArgument := strings.Cut(strings.TrimSpace(part), "=")
			name = strings.ToLower(strings.TrimSpace(name))
			argument = strings.Trim(strings.TrimSpace(argument), `"`)

			switch name {
			case "max-age":
				if hasArgument {
					if age, ok := cacheDirectiveDuration(argument); ok {
						directives.hasMaxAge = true
						directives.maxAge = age
					}
				}
			case "stale-while-revalidate":
				if hasArgument {
					if age, ok := cacheDirectiveDuration(argument); ok {
						directives.hasStaleWhileRevalidate = true
						directives.staleWhileRevalidate = age
					}
				}
			case "no-cache":
				directives.noCache = true
			case "no-store":
				directives.noStore = true
			}
		}
	}
	return directives
}

func splitCacheControl(value string) []string {
	var parts []string
	start := 0
	quoted := false
	escaped := false
	for index, char := range value {
		if escaped {
			escaped = false
			continue
		}
		if quoted && char == '\\' {
			escaped = true
			continue
		}
		if char == '"' {
			quoted = !quoted
			continue
		}
		if char == ',' && !quoted {
			parts = append(parts, value[start:index])
			start = index + 1
		}
	}
	return append(parts, value[start:])
}

func cacheDirectiveDuration(value string) (time.Duration, bool) {
	seconds, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, false
	}
	const maximumSeconds = uint64(1<<63-1) / uint64(time.Second)
	if seconds > maximumSeconds {
		return time.Duration(1<<63 - 1), true
	}
	return time.Duration(seconds) * time.Second, true
}

func cloneResponse(response Response) Response {
	response.Articles = slices.Clone(response.Articles)
	for index := range response.Articles {
		response.Articles[index].Terms = slices.Clone(response.Articles[index].Terms)
	}
	return response
}

func (c *Client) fetchArticles(ctx context.Context, endpoint string) (Response, error) {
	for {
		cached, hasCached := c.cache.get(endpoint)
		now := time.Now()
		if hasCached && cached.isFresh(now) {
			return cloneResponse(cached.response), nil
		}
		if hasCached && cached.canServeStale(now) {
			c.revalidateInBackground(endpoint, cached)
			return cloneResponse(cached.response), nil
		}

		done, owner := c.cache.beginRevalidation(endpoint)
		if !owner {
			select {
			case <-done:
				continue
			case <-ctx.Done():
				return Response{}, fmt.Errorf("buscar notícias: %w", ctx.Err())
			}
		}

		response, updated, shouldStore, err := c.requestArticles(ctx, endpoint, cached, hasCached)
		if err == nil {
			if shouldStore {
				c.cache.store(endpoint, updated)
			} else {
				c.cache.delete(endpoint)
			}
		}
		c.cache.finishRevalidation(endpoint)
		return response, err
	}
}

func (c *Client) revalidateInBackground(endpoint string, cached cachedAPIResponse) {
	_, owner := c.cache.beginRevalidation(endpoint)
	if !owner {
		return
	}

	go func() {
		defer c.cache.finishRevalidation(endpoint)
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		_, updated, shouldStore, err := c.requestArticles(ctx, endpoint, cached, true)
		if err != nil {
			return
		}
		if shouldStore {
			c.cache.store(endpoint, updated)
		} else {
			c.cache.delete(endpoint)
		}
	}()
}

func (c *Client) requestArticles(ctx context.Context, endpoint string, cached cachedAPIResponse, hasCached bool) (result Response, cacheEntry cachedAPIResponse, shouldStore bool, resultErr error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Response{}, cachedAPIResponse{}, false, fmt.Errorf("criar requisição de notícias: %w", err)
	}
	if hasCached && cached.entityTag != "" {
		req.Header.Set("If-None-Match", cached.entityTag)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, cachedAPIResponse{}, false, fmt.Errorf("buscar notícias: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && resultErr == nil {
			result = Response{}
			cacheEntry = cachedAPIResponse{}
			shouldStore = false
			resultErr = fmt.Errorf("fechar resposta da API de notícias: %w", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusNotModified {
		if !hasCached {
			return Response{}, cachedAPIResponse{}, false, errors.New("API de notícias respondeu 304 sem uma resposta em cache")
		}
		updated := cached
		if entityTag := resp.Header.Get("ETag"); entityTag != "" {
			updated.entityTag = entityTag
		}
		if directives := parseCacheDirectives(resp.Header.Values("Cache-Control")); directives.present {
			updated.maxAge = directives.maxAge
			updated.staleWhileRevalidate = directives.staleWhileRevalidate
			updated.noCache = directives.noCache
			updated.receivedAt = time.Now()
			updated.response.FetchedAt = updated.receivedAt
			return cloneResponse(updated.response), updated, !directives.noStore, nil
		}
		updated.receivedAt = time.Now()
		updated.response.FetchedAt = updated.receivedAt
		return cloneResponse(updated.response), updated, true, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maximumErrorResponseLength))
		if readErr == nil {
			if detail := apiErrorMessage(body); detail != "" {
				return Response{}, cachedAPIResponse{}, false, fmt.Errorf("API de notícias respondeu %s: %s", resp.Status, detail)
			}
		}
		return Response{}, cachedAPIResponse{}, false, fmt.Errorf("API de notícias respondeu %s", resp.Status)
	}

	response, err := decodeJSON[Response](resp.Body)
	if err != nil {
		return Response{}, cachedAPIResponse{}, false, fmt.Errorf("decodificar resposta da API de notícias: %w", err)
	}
	if response.Articles == nil {
		response.Articles = []Article{}
	}
	slices.SortStableFunc(response.Articles, func(a, b Article) int {
		return b.Date.Compare(a.Date)
	})

	directives := parseCacheDirectives(resp.Header.Values("Cache-Control"))
	receivedAt := time.Now()
	response.FetchedAt = receivedAt
	updated := cachedAPIResponse{
		response:             cloneResponse(response),
		entityTag:            resp.Header.Get("ETag"),
		receivedAt:           receivedAt,
		maxAge:               directives.maxAge,
		staleWhileRevalidate: directives.staleWhileRevalidate,
		noCache:              directives.noCache,
	}
	shouldStore = !directives.noStore && (updated.entityTag != "" || directives.hasMaxAge || directives.hasStaleWhileRevalidate || directives.noCache)
	return response, updated, shouldStore, nil
}
