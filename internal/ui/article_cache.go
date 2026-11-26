package ui

import (
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
)

const articleCacheKey = "articles-cache"

type articleSnapshot struct {
	Articles []news.Article `json:"articles"`
	Total    int            `json:"total"`
	Updated  time.Time      `json:"updated"`
}

func loadArticleSnapshot(preferences fyne.Preferences) articleSnapshot {
	snapshot := articleSnapshot{Articles: []news.Article{}}
	if value := preferences.String(articleCacheKey); value != "" {
		if err := json.Unmarshal([]byte(value), &snapshot); err != nil {
			return articleSnapshot{Articles: []news.Article{}}
		}
	}
	if snapshot.Articles == nil {
		snapshot.Articles = []news.Article{}
	}
	return snapshot
}

func saveArticleSnapshot(preferences fyne.Preferences, snapshot articleSnapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("serializar cache de notícias: %w", err)
	}
	preferences.SetString(articleCacheKey, string(data))
	return nil
}
