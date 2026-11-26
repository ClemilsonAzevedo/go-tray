package app

import (
	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/native"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/ui"
)

// Run initializes and runs the desktop application.
func Run() {
	a := fyneapp.NewWithID("com.clema.looknews")
	a.Settings().SetTheme(ui.NewTheme())
	a.SetIcon(appIcon())

	drv, ok := a.Driver().(desktop.Driver)
	if !ok {
		panic("este app precisa de um driver desktop")
	}

	desk, ok := a.(desktop.App)
	if !ok {
		panic("system tray não suportado nesta plataforma")
	}

	w := drv.CreateSplashWindow()
	w.Resize(ui.PopoverSize())
	w.SetFixedSize(true)
	w.SetPadded(false)
	w.SetContent(ui.Content(a, news.NewClient()))
	w.Canvas().SetOnTypedKey(
		func(e *fyne.KeyEvent) {
			if e.Name == fyne.KeyEscape {
				w.Hide()
			}
		},
	)

	setupTray(a, desk, w)

	a.Lifecycle().SetOnStarted(
		func() {
			w.Show()
			w.Hide()
			native.StartPopoverPositioner(w)
		},
	)
	a.Run()
}
