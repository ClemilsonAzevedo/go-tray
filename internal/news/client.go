package news

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultAPIURL              = "https://look-news-api.onrender.com"
	requestTimeout             = 90 * time.Second
	maximumErrorResponseLength = 1 << 20
)

type apiError struct {
	Detail   string         `json:"detail"`
	Errors   []apiErrorItem `json:"errors"`
	Instance string         `json:"instance"`
	Status   int            `json:"status"`
	Title    string         `json:"title"`
	Type     string         `json:"type"`
}

type apiErrorItem struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// Client calls the Look News API and validates RSS feed URLs.
type Client struct {
	baseURL    string
	httpClient *http.Client
	cache      httpResponseCache
}

var _ Service = (*Client)(nil)

// NewClient creates an API client using the configured backend.
func NewClient() *Client {
	baseURL := strings.TrimSpace(os.Getenv("LOOK_NEWS_BACKEND_URL"))
	if baseURL == "" {
		baseURL = defaultAPIURL
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// FetchArticles retrieves and sorts articles from the configured RSS sources.
func (c *Client) FetchArticles(ctx context.Context, sources []string) (Response, error) {
	if len(sources) == 0 {
		return Response{}, errors.New("nenhuma fonte RSS foi configurada")
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return Response{}, fmt.Errorf("analisar URL da API: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return Response{}, fmt.Errorf("URL da API inválida: %q", c.baseURL)
	}

	endpoint := baseURL.JoinPath("news")
	query := endpoint.Query()
	for _, source := range sources {
		query.Add("sources", source)
	}
	endpoint.RawQuery = query.Encode()

	requestContext, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	return c.fetchArticles(requestContext, endpoint.String())
}

func apiErrorMessage(body []byte) string {
	response, err := decodeJSON[apiError](bytes.NewReader(body))
	if err == nil {
		if message := response.message(); message != "" {
			return message
		}
	}

	responses, err := decodeJSON[[]apiError](bytes.NewReader(body))
	if err != nil {
		return ""
	}
	messages := make([]string, 0, len(responses))
	for _, response := range responses {
		if message := response.message(); message != "" {
			messages = append(messages, message)
		}
	}
	return strings.Join(messages, "; ")
}

func (e apiError) message() string {
	if e.Detail != "" {
		return e.Detail
	}
	if len(e.Errors) > 0 {
		messages := make([]string, 0, len(e.Errors))
		for _, item := range e.Errors {
			if item.Reason != "" {
				messages = append(messages, item.Reason)
			}
		}
		if len(messages) > 0 {
			return strings.Join(messages, "; ")
		}
	}
	return e.Title
}

func decodeJSON[T any](reader io.Reader) (T, error) {
	var value T
	if err := json.NewDecoder(reader).Decode(&value); err != nil {
		return value, err
	}
	return value, nil
}
