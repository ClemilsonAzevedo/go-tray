package ui

import "fyne.io/fyne/v2"

func showAboutWindow(app fyne.App) {
	if aboutWin != nil {
		focusWindow(aboutWin)
		return
	}

	window := app.NewWindow(appName)
	aboutWin = window
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(320, 280))
	window.SetOnClosed(func() {
		aboutWin = nil
		aboutWindowView = nil
	})
	view := aboutContent(app)
	aboutWindowView = view
	window.SetContent(view.content)
	focusWindow(window)
}
