package news

import "time"

// Article is a news article returned by the Look News API.
type Article struct {
	Author    string    `json:"author"`
	Date      time.Time `json:"date"`
	Link      string    `json:"link"`
	Published string    `json:"published"`
	Source    string    `json:"source"`
	Summary   string    `json:"summary"`
	Terms     []string  `json:"terms"`
	Title     string    `json:"title"`
}

// Response contains the articles and result count returned by the API.
type Response struct {
	Articles []Article `json:"articles"`
	Total    int       `json:"total"`
}
