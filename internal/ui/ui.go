package ui

import (
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

// PopoverSize returns the fixed size used by the tray window.
func PopoverSize() fyne.Size {
	return fyne.NewSize(popoverWidth, popoverHeight)
}

func iconButton(res fyne.Resource, fn func()) *widget.Button {
	b := widget.NewButtonWithIcon("", res, fn)
	b.Importance = widget.LowImportance
	return b
}

func muted(text string, size float32, bold bool) *canvas.Text {
	t := canvas.NewText(text, textMuted())
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return t
}

func metadataText(text string, size float32, bold bool, colorName fyne.ThemeColorName) *canvas.Text {
	t := canvas.NewText(text, systemColor(colorName))
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
