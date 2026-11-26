package ui

import (
	"fyne.io/fyne/v2"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
)

func showSourcesWindow(app fyne.App, client news.Service, onChange func()) {
	if sourcesWin != nil {
		focusWindow(sourcesWin)
		return
	}

	window := app.NewWindow(appName)
	sourcesWin = window
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(380, 320))
	window.SetOnClosed(func() {
		sourcesWin = nil
		sourcesWindowView = nil
	})

	view := newSourcesView(app, client, onChange)
	sourcesWindowView = view
	window.SetContent(view.content())
	view.rebuildSources()
	focusWindow(window)
}
