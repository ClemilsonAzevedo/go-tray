package ui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
)

const (
	popoverWidth  = 380
	popoverHeight = 520
)

// Content creates the main news panel for the app.
func Content(a fyne.App, service news.Service) *fyne.Container {
	return newNewsContent(a, service)
}

func observeTheme(a fyne.App, onChange func(fyne.ThemeVariant)) {
	var mu sync.Mutex
	current := a.Settings().ThemeVariant()
	a.Settings().AddListener(func(settings fyne.Settings) {
		variant := settings.ThemeVariant()
		mu.Lock()
		changed := current != variant
		current = variant
		mu.Unlock()
		if !changed {
			return
		}

		fyne.Do(func() { onChange(a.Settings().ThemeVariant()) })
	})
}

// PopoverSize returns the fixed size used by the tray window.
func PopoverSize() fyne.Size {
	return fyne.NewSize(popoverWidth, popoverHeight)
}

func iconButton(res fyne.Resource, fn func()) *widget.Button {
	b := widget.NewButtonWithIcon("", res, fn)
	b.Importance = widget.LowImportance
	return b
}

func metadataText(text string, size float32, bold bool, colorName fyne.ThemeColorName) *canvas.Text {
	return metadataTextForVariant(text, size, bold, colorName, currentThemeVariant())
}

func metadataTextForVariant(text string, size float32, bold bool, colorName fyne.ThemeColorName, variant fyne.ThemeVariant) *canvas.Text {
	t := canvas.NewText(text, systemColorForVariant(colorName, variant))
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return t
}

func richText(text string, colorName fyne.ThemeColorName, sizeName fyne.ThemeSizeName, bold bool) *widget.RichText {
	r := widget.NewRichText(
		&widget.TextSegment{
			Text: text,
			Style: widget.RichTextStyle{
				ColorName: colorName,
				SizeName:  sizeName,
				TextStyle: fyne.TextStyle{Bold: bold},
			},
		},
	)
	r.Wrapping = fyne.TextWrapWord
	return r
}
