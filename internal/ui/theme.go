package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type lookTheme struct{}

const (
	articleAccentColorName            fyne.ThemeColorName = "lookNewsArticleAccent"
	articleSummaryColorName           fyne.ThemeColorName = "lookNewsArticleSummary"
	articleMetadataSecondaryColorName fyne.ThemeColorName = "lookNewsMetadataSecondary"
	articleMetadataTertiaryColorName  fyne.ThemeColorName = "lookNewsMetadataTertiary"
)

// NewTheme returns the system-adaptive Fyne theme with slightly roomier spacing.
func NewTheme() fyne.Theme { return lookTheme{} }

func (lookTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0xED, G: 0xB2, B: 0x99, A: 0xFF}
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0x2B, G: 0x22, B: 0x1E, A: 0xFF}
	case theme.ColorNameShadow:
		return color.Transparent
	case theme.ColorNameScrollBar:
		return color.Transparent
	case articleAccentColorName:
		return color.NRGBA{R: 0xED, G: 0xB2, B: 0x99, A: 0xFF}
	case articleSummaryColorName, articleMetadataSecondaryColorName:
		if v == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x99}
		}
		return color.NRGBA{A: 0x8C}
	case articleMetadataTertiaryColorName:
		if v == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x61}
		}
		return color.NRGBA{A: 0x59}
	}
	return theme.DefaultTheme().Color(n, v)
}

func (lookTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(s)
}

func (lookTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (lookTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNameText:
		return 12
	case theme.SizeNamePadding:
		return 7
	case theme.SizeNameInnerPadding:
		return 7
	}
	return theme.DefaultTheme().Size(n)
}

func systemColor(name fyne.ThemeColorName) color.Color {
	app := fyne.CurrentApp()
	if app == nil {
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
	settings := app.Settings()
	return settings.Theme().Color(name, settings.ThemeVariant())
}

func textMain() color.Color { return systemColor(theme.ColorNameForeground) }

func textMuted() color.Color { return systemColor(theme.ColorNameDisabled) }

func cardColor() color.Color { return systemColor(theme.ColorNameBackground) }

func cardHoverColor() color.Color {
	return color.NRGBA{R: 0xE9, G: 0xAD, B: 0x8D, A: 0x24}
}

func lineColor() color.Color { return systemColor(theme.ColorNameInputBorder) }
