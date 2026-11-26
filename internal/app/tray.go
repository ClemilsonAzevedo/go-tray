package app

import (
	_ "embed"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/native"
)

//go:embed assets/icon.png
var iconPNG []byte

//go:embed assets/trayIconTemplate.png
var trayIconTemplatePNG []byte

func appIcon() fyne.Resource {
	return fyne.NewStaticResource("icon.png", iconPNG)
}

func trayIcon() fyne.Resource {
	if runtime.GOOS == "darwin" {
		return fyne.NewStaticResource("trayIconTemplate.png", trayIconTemplatePNG)
	}
	return appIcon()
}

// setupTray configures the tray icon, popup and quit menu.
func setupTray(a fyne.App, desk desktop.App, w fyne.Window) {
	desk.SetSystemTrayIcon(trayIcon())
	desk.SetSystemTrayWindow(w)

	quit := fyne.NewMenuItem("Sair do Look News", func() { a.Quit() })
	quit.IsQuit = true
	desk.SetSystemTrayMenu(fyne.NewMenu("Look News", quit))

	// The status item must exist before its image can be marked as a template.
	if runtime.GOOS == "darwin" {
		go func() {
			time.Sleep(time.Second)
			fyne.DoAndWait(native.MarkTrayIconTemplate)
		}()
	}
}
