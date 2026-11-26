package ui

import (
	"fyne.io/fyne/v2"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/native"
)

const (
	appName    = "look news"
	appVersion = "2026.11.26"
	sourcesKey = "sources"
)

var sourcesWin, aboutWin fyne.Window

func loadSources(a fyne.App) []string {
	return a.Preferences().StringList(sourcesKey)
}

func focusWindow(w fyne.Window) {
	w.Show()
	w.RequestFocus()
	native.ActivateApp()
}
