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
	"slices"
	"strings"
	"time"
)

const (
	defaultAPIURL              = "https://look-news-api.onrender.com"
	defaultCriterion           = "Ciência, tecnologia, startups e investimentos"
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
	criterion  string
	httpClient *http.Client
}

var _ Service = (*Client)(nil)

// NewClient creates an API client using the configured backend and criterion.
func NewClient() *Client {
	baseURL := strings.TrimSpace(os.Getenv("LOOK_NEWS_BACKEND_URL"))
	if baseURL == "" {
		baseURL = defaultAPIURL
	}

	criterion := strings.TrimSpace(os.Getenv("LOOK_NEWS_CRITERION"))
	if criterion == "" {
		criterion = defaultCriterion
	}

	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		criterion: criterion,
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
	query.Set("criterion", c.criterion)
	for _, source := range sources {
		query.Add("sources", source)
	}
	endpoint.RawQuery = query.Encode()

	requestContext, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestContext, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return Response{}, fmt.Errorf("criar requisição de notícias: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("buscar notícias: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, maximumErrorResponseLength))
		if err == nil {
			if detail := apiErrorMessage(body); detail != "" {
				return Response{}, fmt.Errorf("API de notícias respondeu %s: %s", resp.Status, detail)
			}
		}
		return Response{}, fmt.Errorf("API de notícias respondeu %s", resp.Status)
	}

	response, err := decodeJSON[Response](resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("decodificar resposta da API de notícias: %w", err)
	}
	if response.Articles == nil {
		response.Articles = []Article{}
	}
	slices.SortStableFunc(response.Articles, func(a, b Article) int {
		return b.Date.Compare(a.Date)
	})
	return response, nil
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
