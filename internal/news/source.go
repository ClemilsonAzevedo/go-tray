package news

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const sourceValidationTimeout = 4 * time.Second

var feedRootPattern = regexp.MustCompile(`(?i)<(?:rss[\s>]|feed[\s>]|rdf:rdf[\s>])`)

// NormalizeSourceURL validates and normalizes an HTTP or HTTPS feed URL.
func NormalizeSourceURL(raw string) (string, error) {
	value := strings.Trim(strings.TrimSpace(raw), `<>'"`)
	if value == "" {
		return "", errors.New("Digite uma URL.")
	}

	if strings.ContainsAny(value, " \t\r\n") {
		return "", errors.New("A URL não pode ter espaços.")
	}

	if !strings.Contains(value, "://") {
		value = "https://" + value
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return "", errors.New("URL inválida.")
	}

	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil {
		return "", errors.New("Só são aceitas URLs http:// ou https://.")
	}

	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.String(), nil
}

// ValidateFeed checks that a URL responds with an RSS or Atom feed.
func (c *Client) ValidateFeed(ctx context.Context, source string) error {
	return validateRSSFeed(ctx, c.httpClient, source)
}

func validateRSSFeed(ctx context.Context, client *http.Client, source string) error {
	requestContext, cancel := context.WithTimeout(ctx, sourceValidationTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestContext, http.MethodGet, source, nil)
	if err != nil {
		return fmt.Errorf("criar requisição para validar feed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("não foi possível acessar essa URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("A URL respondeu %d.", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2_000))
	if err != nil {
		return fmt.Errorf("ler resposta do feed: %w", err)
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "xml") &&
		!strings.Contains(contentType, "rss") &&
		!strings.Contains(contentType, "atom") &&
		!feedRootPattern.Match(body) {

		return errors.New("Essa URL não parece ser um feed RSS/Atom.")
	}

	return nil
}