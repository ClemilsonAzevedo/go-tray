package ui

import (
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fyneTest "fyne.io/fyne/v2/test"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
)

func TestSameArticles(t *testing.T) {
	t.Parallel()

	article := news.Article{
		Author:    "Author",
		Date:      time.Date(2026, time.November, 25, 12, 0, 0, 0, time.UTC),
		Link:      "https://example.com/article",
		Published: "Wed, 25 Nov 2026 12:00:00 +0000",
		Source:    "Example",
		Summary:   "Summary",
		Terms:     []string{"Go", "macOS"},
		Title:     "Article title",
	}

	if !sameArticles([]news.Article{article}, []news.Article{article}) {
		t.Fatal("identical article lists should compare equal")
	}
	if !sameArticles(nil, []news.Article{}) {
		t.Fatal("empty article lists should compare equal")
	}

	changed := article
	changed.Terms = []string{"Go"}
	if sameArticles([]news.Article{article}, []news.Article{changed}) {
		t.Fatal("article lists with different terms should not compare equal")
	}
}

func TestListTopShadowTracksScrollPosition(t *testing.T) {
	shadow := newListTopShadow()
	if got, want := color.NRGBAModel.Convert(shadow.StartColor).(color.NRGBA), (color.NRGBA{}); got != want {
		t.Fatalf("initial top shadow color = %v, want transparent", got)
	}

	updateListTopShadow(shadow, fyne.NewPos(0, 1))
	if got, want := color.NRGBAModel.Convert(shadow.StartColor).(color.NRGBA), (color.NRGBA{A: 0x24}); got != want {
		t.Fatalf("scrolled top shadow color = %v, want %v", got, want)
	}

	updateListTopShadow(shadow, fyne.Position{})
	if got, want := color.NRGBAModel.Convert(shadow.StartColor).(color.NRGBA), (color.NRGBA{}); got != want {
		t.Fatalf("top shadow color at scroll origin = %v, want transparent", got)
	}
}

func TestShowMessageInvalidatesRenderedArticles(t *testing.T) {
	t.Parallel()
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)

	article := news.Article{Title: "Previously rendered"}
	view := &newsView{
		list:     container.NewVBox(),
		rendered: []news.Article{article},
	}
	view.showMessage("No articles")

	if sameArticles(view.rendered, []news.Article{article}) {
		t.Fatal("a message should invalidate the previously rendered articles")
	}
}
