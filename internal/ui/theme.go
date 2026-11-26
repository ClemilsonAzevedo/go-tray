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
	statusMetadataColorName           fyne.ThemeColorName = "lookNewsStatusMetadata"
)

// NewTheme returns the system-adaptive Fyne theme with slightly roomier spacing.
func NewTheme() fyne.Theme { return lookTheme{} }

func themeIcon(variant fyne.ThemeVariant, light, dark fyne.Resource) fyne.Resource {
	if variant == theme.VariantLight {
		return light
	}
	return dark
}

func (lookTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNamePrimary:
		if v == theme.VariantLight {
			return color.NRGBA{R: 0xE8, G: 0xAA, B: 0x90, A: 0xFF}
		}
		return color.NRGBA{R: 0xED, G: 0xB2, B: 0x99, A: 0xFF}
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0x2B, G: 0x22, B: 0x1E, A: 0xFF}
	case theme.ColorNameShadow:
		return color.Transparent
	case theme.ColorNameScrollBar:
		return color.Transparent
	case articleAccentColorName:
		if v == theme.VariantLight {
			return color.NRGBA{R: 0xE8, G: 0xAA, B: 0x90, A: 0xFF}
		}
		return color.NRGBA{R: 0xED, G: 0xB2, B: 0x99, A: 0xFF}
	case articleSummaryColorName:
		if v == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x99}
		}
		return color.NRGBA{A: 0x8C}
	case articleMetadataSecondaryColorName:
		if v == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x99}
		}
		return color.NRGBA{R: 0x63, G: 0x5B, B: 0x55, A: 0xFF}
	case articleMetadataTertiaryColorName:
		if v == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x61}
		}
		return color.NRGBA{A: 0x59}
	case statusMetadataColorName:
		if v == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x61}
		}
		return color.NRGBA{R: 0x78, G: 0x70, B: 0x6A, A: 0xFF}
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
	return systemColorForVariant(name, currentThemeVariant())
}

func currentThemeVariant() fyne.ThemeVariant {
	if app := fyne.CurrentApp(); app != nil {
		return app.Settings().ThemeVariant()
	}
	return theme.VariantLight
}

func systemColorForVariant(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	app := fyne.CurrentApp()
	if app == nil {
		return theme.DefaultTheme().Color(name, variant)
	}
	return app.Settings().Theme().Color(name, variant)
}

func textMain() color.Color { return systemColor(theme.ColorNameForeground) }

func cardColor() color.Color { return systemColor(theme.ColorNameBackground) }

func cardHoverColor() color.Color {
	return color.NRGBA{R: 0xE9, G: 0xAD, B: 0x8D, A: 0x24}
}

func lineColor() color.Color { return systemColor(theme.ColorNameInputBorder) }
