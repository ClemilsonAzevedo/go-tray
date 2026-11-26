package news

import "context"

// Service retrieves news articles and validates RSS sources.
type Service interface {
	FetchArticles(context.Context, []string) (Response, error)
	ValidateFeed(context.Context, string) error
}
