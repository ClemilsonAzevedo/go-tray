package ui

import (
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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

func TestNewsViewRefreshThemeRebuildsExistingArticles(t *testing.T) {
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)
	app.Settings().SetTheme(NewTheme())

	article := news.Article{Source: "Example", Title: "Story", Summary: "Description"}
	view := &newsView{
		app:        app,
		content:    container.NewVBox(),
		list:       container.NewVBox(),
		title:      canvas.NewText("look news", color.Black),
		brandIcon:  widget.NewIcon(icEye),
		count:      canvas.NewText("1 artigo", color.Black),
		footerDot:  canvas.NewText("·", color.Black),
		updated:    canvas.NewText("atualizado agora", color.Black),
		infoButton: widget.NewButtonWithIcon("", icInfo, nil),
		separators: []*canvas.Rectangle{canvas.NewRectangle(color.Black)},
		topShadow:  newListTopShadow(),
	}
	view.showArticlesForVariant([]news.Article{article}, theme.VariantLight)
	oldRow := view.list.Objects[0]

	view.refreshTheme(theme.VariantDark)

	if view.list.Objects[0] == oldRow {
		t.Fatal("theme refresh should rebuild existing article cards")
	}
	if got, want := view.brandIcon.Resource.Name(), icEye.Name(); got != want {
		t.Errorf("brand icon = %q, want %q", got, want)
	}
	if got, want := view.infoButton.Icon.Name(), icInfo.Name(); got != want {
		t.Errorf("about icon = %q, want %q", got, want)
	}

	source := findCanvasText(view.list.Objects[0], "EXAMPLE")
	if source == nil {
		t.Fatal("rebuilt article source text was not found")
	}
	wantColor := color.NRGBAModel.Convert(NewTheme().Color(articleAccentColorName, theme.VariantDark)).(color.NRGBA)
	if got := color.NRGBAModel.Convert(source.Color).(color.NRGBA); got != wantColor {
		t.Errorf("article source color = %v, want %v", got, wantColor)
	}
	if !sameArticles(view.rendered, []news.Article{article}) {
		t.Errorf("theme refresh changed rendered article data: %v", view.rendered)
	}
}

func findCanvasText(object fyne.CanvasObject, text string) *canvas.Text {
	if current, ok := object.(*canvas.Text); ok && current.Text == text {
		return current
	}
	if current, ok := object.(*fyne.Container); ok {
		for _, child := range current.Objects {
			if found := findCanvasText(child, text); found != nil {
				return found
			}
		}
	}
	if current, ok := object.(*tappable); ok {
		return findCanvasText(current.content, text)
	}
	return nil
}
