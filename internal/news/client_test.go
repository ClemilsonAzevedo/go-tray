package news

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestDecodeNewsResponse(t *testing.T) {
	t.Parallel()

	const body = `{"articles":[{"author":"Marina Temkin","date":"2026-11-25T22:44:32Z","link":"https://techcrunch.com/2026/11/25/example/","published":"Wed, 25 Nov 2026 22:44:32 +0000","source":"TechCrunch","summary":"A sample summary.","terms":["science"],"title":"A sample headline"}],"total":1}`

	var response Response
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("decode API response: %v", err)
	}
	if response.Total != 1 || len(response.Articles) != 1 {
		t.Fatalf("unexpected response totals: total=%d articles=%d", response.Total, len(response.Articles))
	}
	article := response.Articles[0]
	if article.Title != "A sample headline" || article.Published != "Wed, 25 Nov 2026 22:44:32 +0000" {
		t.Fatalf("unexpected article fields: %#v", article)
	}
	if want := time.Date(2026, time.November, 25, 22, 44, 32, 0, time.UTC); !article.Date.Equal(want) {
		t.Fatalf("article date = %s, want %s", article.Date, want)
	}
}

func TestNewsClientUsesSourcesQueryParameter(t *testing.T) {
	t.Parallel()

	sources := []string{"https://one.example/feed.xml", "https://two.example/rss"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/news" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if _, ok := r.URL.Query()["criterion"]; ok {
			t.Errorf("request unexpectedly includes criterion: %s", r.URL.RawQuery)
		}
		if got := r.URL.Query()["sources"]; !slices.Equal(got, sources) {
			t.Errorf("sources = %#v, want %#v", got, sources)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"articles":[],"total":0}`))
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	response, err := client.FetchArticles(context.Background(), sources)
	if err != nil {
		t.Fatalf("fetch articles: %v", err)
	}
	if response.Total != 0 || response.Articles == nil {
		t.Fatalf("unexpected empty response: %#v", response)
	}
}

func TestNewsClientReusesFreshCachedResponse(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Cache-Control", "max-age=60, stale-while-revalidate=300")
		w.Header().Set("ETag", `"news-v1"`)
		_, _ = w.Write([]byte(`{"articles":[{"title":"cached"}],"total":1}`))
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	sources := []string{"https://one.example/feed.xml"}
	var fetchedAt time.Time
	for range 2 {
		response, err := client.FetchArticles(context.Background(), sources)
		if err != nil {
			t.Fatalf("fetch articles: %v", err)
		}
		if got := response.Articles[0].Title; got != "cached" {
			t.Fatalf("article title = %q, want cached", got)
		}
		if response.FetchedAt.IsZero() {
			t.Fatal("cached response has no fetch timestamp")
		}
		if fetchedAt.IsZero() {
			fetchedAt = response.FetchedAt
		} else if !response.FetchedAt.Equal(fetchedAt) {
			t.Fatalf("fresh cache timestamp = %s, want original fetch time %s", response.FetchedAt, fetchedAt)
		}
	}
	if requests != 1 {
		t.Fatalf("request count = %d, want 1 while the response is fresh", requests)
	}
}

func TestNewsClientRevalidatesCachedResponseWithETag(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			w.Header().Set("Cache-Control", "max-age=0")
			w.Header().Set("ETag", `"news-v1"`)
			_, _ = w.Write([]byte(`{"articles":[{"title":"cached"}],"total":1}`))
			return
		}

		if got := r.Header.Get("If-None-Match"); got != `"news-v1"` {
			t.Errorf("If-None-Match = %q, want %q", got, `"news-v1"`)
		}
		w.Header().Set("Cache-Control", "max-age=60")
		w.Header().Set("ETag", `"news-v1"`)
		w.WriteHeader(http.StatusNotModified)
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	sources := []string{"https://one.example/feed.xml"}
	var firstFetchedAt time.Time
	for index := range 2 {
		response, err := client.FetchArticles(context.Background(), sources)
		if err != nil {
			t.Fatalf("fetch articles: %v", err)
		}
		if got := response.Articles[0].Title; got != "cached" {
			t.Fatalf("article title = %q, want cached", got)
		}
		if index == 0 {
			firstFetchedAt = response.FetchedAt
		} else if !response.FetchedAt.After(firstFetchedAt) {
			t.Fatalf("304 validation time = %s, want after initial response time %s", response.FetchedAt, firstFetchedAt)
		}
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want initial request and one conditional revalidation", requests)
	}
}

func TestNewsClientRevalidatesStaleResponseInBackground(t *testing.T) {
	t.Parallel()

	requests := 0
	revalidationStarted := make(chan struct{})
	releaseRevalidation := make(chan struct{})
	revalidationResponded := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			w.Header().Set("Cache-Control", "max-age=0, stale-while-revalidate=60")
			w.Header().Set("ETag", `"news-v1"`)
			_, _ = w.Write([]byte(`{"articles":[{"title":"stale"}],"total":1}`))
			return
		}
		if requests != 2 {
			t.Errorf("unexpected request count %d", requests)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if got := r.Header.Get("If-None-Match"); got != `"news-v1"` {
			t.Errorf("If-None-Match = %q, want %q", got, `"news-v1"`)
		}
		close(revalidationStarted)
		select {
		case <-releaseRevalidation:
		case <-r.Context().Done():
			return
		}
		w.Header().Set("Cache-Control", "max-age=60, stale-while-revalidate=300")
		w.Header().Set("ETag", `"news-v2"`)
		_, _ = w.Write([]byte(`{"articles":[{"title":"fresh"}],"total":1}`))
		close(revalidationResponded)
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	sources := []string{"https://one.example/feed.xml"}
	initial, err := client.FetchArticles(context.Background(), sources)
	if err != nil {
		t.Fatalf("initial fetch: %v", err)
	}
	if got := initial.Articles[0].Title; got != "stale" {
		t.Fatalf("initial title = %q, want stale", got)
	}

	type fetchResult struct {
		response Response
		err      error
	}
	result := make(chan fetchResult, 1)
	go func() {
		response, err := client.FetchArticles(context.Background(), sources)
		result <- fetchResult{response: response, err: err}
	}()

	select {
	case fetched := <-result:
		if fetched.err != nil {
			t.Fatalf("stale fetch: %v", fetched.err)
		}
		if got := fetched.response.Articles[0].Title; got != "stale" {
			t.Fatalf("stale-while-revalidate title = %q, want stale", got)
		}
	case <-time.After(time.Second):
		close(releaseRevalidation)
		t.Fatal("stale response waited for background revalidation")
	}

	select {
	case <-revalidationStarted:
	case <-time.After(time.Second):
		close(releaseRevalidation)
		t.Fatal("background revalidation did not start")
	}
	close(releaseRevalidation)
	select {
	case <-revalidationResponded:
	case <-time.After(time.Second):
		t.Fatal("background revalidation did not finish its response")
	}

	deadline := time.Now().Add(time.Second)
	for {
		client.cache.mu.Lock()
		inFlight := len(client.cache.revalidations) != 0
		client.cache.mu.Unlock()
		if !inFlight {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background revalidation remained in flight")
		}
		time.Sleep(time.Millisecond)
	}

	updated, err := client.FetchArticles(context.Background(), sources)
	if err != nil {
		t.Fatalf("fetch updated response: %v", err)
	}
	if got := updated.Articles[0].Title; got != "fresh" {
		t.Fatalf("updated title = %q, want fresh", got)
	}
	if requests != 2 {
		t.Fatalf("request count = %d, want initial request and background revalidation", requests)
	}
}

func TestNewsClientSortsArticlesNewestFirst(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"articles":[{"date":"2026-01-01T00:00:00Z","title":"older"},{"date":"2026-11-01T00:00:00Z","title":"newer"}],"total":2}`))
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	response, err := client.FetchArticles(context.Background(), []string{"https://one.example/feed.xml"})
	if err != nil {
		t.Fatalf("fetch articles: %v", err)
	}
	if got := response.Articles[0].Title; got != "newer" {
		t.Fatalf("newest article = %q, want %q", got, "newer")
	}
}

func TestNewsClientIncludesHTTPErrorDetail(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"detail":"service is warming up"}`))
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.FetchArticles(context.Background(), []string{"https://one.example/feed.xml"})
	if err == nil || !strings.Contains(err.Error(), "service is warming up") {
		t.Fatalf("error = %v, want API detail", err)
	}
}

func TestNewsClientIncludesArrayHTTPErrorDetail(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`[{"detail":"sources is required"}]`))
	}))
	t.Cleanup(server.Close)

	client := &Client{baseURL: server.URL, httpClient: server.Client()}
	_, err := client.FetchArticles(context.Background(), []string{"https://one.example/feed.xml"})
	if err == nil || !strings.Contains(err.Error(), "sources is required") {
		t.Fatalf("error = %v, want API error-array detail", err)
	}
}

func TestNormalizeSourceURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
		valid bool
	}{
		{input: " <https://example.com/feed.xml> ", want: "https://example.com/feed.xml", valid: true},
		{input: "example.com/feed.xml", want: "https://example.com/feed.xml", valid: true},
		{input: "example.com", want: "https://example.com/", valid: true},
		{input: "ftp://example.com/feed.xml", valid: false},
		{input: "https://example.com/a b", valid: false},
		{input: "", valid: false},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got, err := NormalizeSourceURL(test.input)
			if test.valid {
				if err != nil || got != test.want {
					t.Fatalf("NormalizeSourceURL(%q) = %q, %v; want %q", test.input, got, err, test.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("NormalizeSourceURL(%q) unexpectedly succeeded: %q", test.input, got)
			}
		})
	}
}

func TestValidateRSSFeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		body        string
		status      int
		wantErr     bool
	}{
		{name: "rss content", contentType: "application/xml", body: "<rss version=\"2.0\"></rss>", status: http.StatusOK},
		{name: "atom body", contentType: "text/plain", body: "<feed xmlns=\"http://www.w3.org/2005/Atom\"></feed>", status: http.StatusOK},
		{name: "html page", contentType: "text/html", body: "<html></html>", status: http.StatusOK, wantErr: true},
		{name: "not found", contentType: "text/plain", body: "missing", status: http.StatusNotFound, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", test.contentType)
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			err := validateRSSFeed(context.Background(), server.Client(), server.URL)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateRSSFeed() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}
