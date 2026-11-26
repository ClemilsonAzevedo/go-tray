package ui

import (
	"context"
	"fmt"
	"image/color"
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
)

type newsView struct {
	app           fyne.App
	client        news.Service
	list          *fyne.Container
	count         *canvas.Text
	updated       *canvas.Text
	refreshButton *toolbarButton
	rendered      []news.Article
	lastUpdated   time.Time
	lastAttempt   time.Time
	loading       bool
}

func newNewsContent(a fyne.App, client news.Service) *fyne.Container {
	view := &newsView{
		app:     a,
		client:  client,
		list:    container.New(layout.NewCustomPaddedVBoxLayout(4)),
		count:   metadataText("0 artigos", 11, false, articleMetadataSecondaryColorName),
		updated: metadataText("", 11, false, articleMetadataTertiaryColorName),
	}

	title := canvas.NewText("look news", textMain())
	title.TextSize = 14
	title.TextStyle = fyne.TextStyle{Bold: true}

	view.refreshButton = newToolbarButton(icRefresh, false, view.refresh)
	view.refreshButton.setIconFrames(refreshIconFrames())
	brand := container.New(layout.NewCustomPaddedHBoxLayout(6),
		container.New(layout.NewCustomPaddedLayout(4, 0, 0, 0),
			container.NewGridWrap(fyne.NewSize(20, 20), widget.NewIcon(icEye)),
		),
		title,
	)
	actions := container.New(layout.NewCustomPaddedHBoxLayout(8),
		view.refreshButton,
		newToolbarButton(icLink, false, func() { showSourcesWindow(a, client, view.refresh) }),
		newToolbarButton(icPower, true, a.Quit),
	)
	header := container.New(layout.NewCustomPaddedLayout(8, 8, 14, 14),
		container.New(layout.NewCustomPaddedHBoxLayout(0), brand, layout.NewSpacer(), actions),
	)

	footer := container.New(layout.NewCustomPaddedLayout(6, 6, 14, 14),
		container.New(layout.NewCustomPaddedHBoxLayout(10),
			view.count,
			metadataText("·", 11, false, articleMetadataTertiaryColorName),
			view.updated,
			layout.NewSpacer(),
			iconButton(icInfo, func() { showAboutWindow(a) }),
		),
	)
	separator := func() fyne.CanvasObject {
		line := canvas.NewRectangle(lineColor())
		line.SetMinSize(fyne.NewSize(1, 1))
		return line
	}
	scroll := container.NewVScroll(container.New(layout.NewCustomPaddedLayout(4, 4, 6, 6), view.list))
	topShadow := newListTopShadow()
	scroll.OnScrolled = func(offset fyne.Position) {
		updateListTopShadow(topShadow, offset)
	}
	scrolledList := container.NewStack(
		scroll,
		container.NewBorder(topShadow, nil, nil, nil, nil),
	)

	content := container.NewBorder(
		container.NewVBox(header, separator()),
		container.NewVBox(separator(), footer),
		nil,
		nil,
		scrolledList,
	)
	view.refresh()
	view.startPolling()
	return content
}

func (v *newsView) refresh() {
	if v.loading {
		return
	}

	sources := loadSources(v.app)
	snapshot := loadArticleSnapshot(v.app.Preferences())
	if len(sources) == 0 {
		v.render(snapshot, nil)
		return
	}

	v.loading = true
	v.refreshButton.setLoading(true)
	v.updated.Text = "atualizando..."
	v.updated.Refresh()
	if len(snapshot.Articles) == 0 {
		v.showMessage("Carregando notícias…")
	} else {
		v.showArticles(snapshot.Articles)
		v.count.Text = articleCount(snapshot.Total)
		v.count.Refresh()
	}

	go func() {
		response, err := v.client.FetchArticles(context.Background(), sources)
		if err == nil {
			snapshot = articleSnapshot{
				Articles: response.Articles,
				Total:    response.Total,
				Updated:  time.Now(),
			}
		}
		fyne.Do(func() {
			v.loading = false
			v.lastAttempt = time.Now()
			v.refreshButton.setLoading(false)
			if err == nil {
				if cacheErr := saveArticleSnapshot(v.app.Preferences(), snapshot); cacheErr != nil {
					slog.Error("salvar cache de notícias", "error", cacheErr)
				}
			}
			v.render(snapshot, err)
		})
	}()
}

func newListTopShadow() *canvas.LinearGradient {
	shadow := canvas.NewVerticalGradient(color.Transparent, color.Transparent)
	shadow.SetMinSize(fyne.NewSize(0, 8))
	return shadow
}

func updateListTopShadow(shadow *canvas.LinearGradient, offset fyne.Position) {
	if offset.Y > 0 {
		shadow.StartColor = color.NRGBA{A: 0x24}
	} else {
		shadow.StartColor = color.Transparent
	}
	shadow.Refresh()
}

func (v *newsView) render(snapshot articleSnapshot, err error) {
	v.count.Text = articleCount(snapshot.Total)
	v.count.Refresh()
	if err != nil {
		v.lastUpdated = snapshot.Updated
		if len(snapshot.Articles) > 0 {
			v.showArticles(snapshot.Articles)
			v.updated.Text = "erro · usando cache"
		} else {
			v.showMessage("Não foi possível carregar as notícias.\n" + err.Error())
			v.updated.Text = "erro ao atualizar"
		}
		v.updated.Refresh()
		return
	}

	if len(snapshot.Articles) == 0 {
		if len(loadSources(v.app)) == 0 {
			v.showMessage("Adicione uma fonte RSS para começar.")
		} else {
			v.showMessage("Nenhum artigo por aqui ainda.")
		}
	} else {
		v.showArticles(snapshot.Articles)
	}
	v.lastUpdated = snapshot.Updated
	v.updateUpdatedLabel()
}

func (v *newsView) showArticles(articles []news.Article) {
	if sameArticles(v.rendered, articles) {
		return
	}
	objects := make([]fyne.CanvasObject, len(articles))
	for index, article := range articles {
		objects[index] = articleRow(article, v.openArticle)
	}
	v.list.Objects = objects
	v.rendered = slices.Clone(articles)
	v.list.Refresh()
}

func sameArticles(left, right []news.Article) bool {
	return slices.EqualFunc(left, right, func(left, right news.Article) bool {
		return left.Author == right.Author &&
			left.Date.Equal(right.Date) &&
			left.Link == right.Link &&
			left.Published == right.Published &&
			left.Source == right.Source &&
			left.Summary == right.Summary &&
			slices.Equal(left.Terms, right.Terms) &&
			left.Title == right.Title
	})
}

func (v *newsView) showMessage(message string) {
	v.rendered = nil
	v.list.Objects = []fyne.CanvasObject{
		container.NewPadded(richText(message, theme.ColorNameDisabled, theme.SizeNameText, false)),
	}
	v.list.Refresh()
}

func (v *newsView) openArticle(rawURL string) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		slog.Warn("link de notícia inválido", "url", rawURL)
		return
	}
	if err := v.app.OpenURL(parsed); err != nil {
		slog.Error("abrir link de notícia", "error", err)
	}
}

func (v *newsView) startPolling() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			fyne.Do(func() {
				v.updateUpdatedLabel()
				if v.lastAttempt.IsZero() || time.Since(v.lastAttempt) >= pollInterval() {
					v.refresh()
				}
			})
		}
	}()
}

func pollInterval() time.Duration {
	const defaultInterval = 5 * time.Minute
	value := strings.TrimSpace(os.Getenv("LOOK_NEWS_POLL_INTERVAL_MS"))
	if value == "" {
		return defaultInterval
	}
	interval, err := time.ParseDuration(value + "ms")
	if err != nil || interval <= 0 {
		return defaultInterval
	}
	return interval
}

func (v *newsView) updateUpdatedLabel() {
	if v.lastUpdated.IsZero() {
		return
	}
	v.updated.Text = "atualizado " + ago(v.lastUpdated)
	v.updated.Refresh()
}

func articleCount(count int) string {
	plural := "s"
	if count == 1 {
		plural = ""
	}
	return fmt.Sprintf("%d artigo%s", count, plural)
}
