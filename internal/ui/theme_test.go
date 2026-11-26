package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestThemeUsesOpaqueSystemBackground(t *testing.T) {
	for _, variant := range []fyne.ThemeVariant{theme.VariantLight, theme.VariantDark} {
		got := NewTheme().Color(theme.ColorNameBackground, variant)
		want := theme.DefaultTheme().Color(theme.ColorNameBackground, variant)

		gotR, gotG, gotB, gotA := got.RGBA()
		wantR, wantG, wantB, wantA := want.RGBA()
		if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
			t.Errorf("background for variant %v = %v, want default theme color %v", variant, got, want)
		}
		if gotA != 0xffff {
			t.Errorf("background for variant %v has alpha %04x, want opaque", variant, gotA)
		}
	}
}

func TestThemeHidesScrollBarsAndBuiltInShadows(t *testing.T) {
	for _, variant := range []fyne.ThemeVariant{theme.VariantLight, theme.VariantDark} {
		for _, name := range []fyne.ThemeColorName{theme.ColorNameShadow, theme.ColorNameScrollBar} {
			_, _, _, alpha := NewTheme().Color(name, variant).RGBA()
			if alpha != 0 {
				t.Errorf("theme color %q alpha for variant %v = %d, want transparent", name, variant, alpha)
			}
		}
	}
}

func TestArticleThemeColors(t *testing.T) {
	for _, variant := range []fyne.ThemeVariant{theme.VariantLight, theme.VariantDark} {
		source := NewTheme().Color(articleAccentColorName, variant)
		expectedPeach := color.NRGBA{R: 0xE8, G: 0xAA, B: 0x90, A: 0xFF}
		if variant == theme.VariantDark {
			expectedPeach = color.NRGBA{R: 0xED, G: 0xB2, B: 0x99, A: 0xFF}
		}
		if got := color.NRGBAModel.Convert(source).(color.NRGBA); got != expectedPeach {
			t.Errorf("article source color for variant %v = %v, want %v", variant, got, expectedPeach)
		}

		summary := NewTheme().Color(articleSummaryColorName, variant)
		secondary := NewTheme().Color(articleMetadataSecondaryColorName, variant)
		expectedSummary := color.NRGBA{A: 0x8C}
		expectedSecondary := color.NRGBA{R: 0x63, G: 0x5B, B: 0x55, A: 0xFF}
		if variant == theme.VariantDark {
			expectedSummary = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x99}
			expectedSecondary = expectedSummary
		}
		if got := color.NRGBAModel.Convert(summary).(color.NRGBA); got != expectedSummary {
			t.Errorf("article summary color for variant %v = %v, want %v", variant, got, expectedSummary)
		}
		if got := color.NRGBAModel.Convert(secondary).(color.NRGBA); got != expectedSecondary {
			t.Errorf("metadata secondary color for variant %v = %v, want %v", variant, got, expectedSecondary)
		}
		tertiary := NewTheme().Color(articleMetadataTertiaryColorName, variant)
		expectedTertiary := color.NRGBA{A: 0x59}
		if variant == theme.VariantDark {
			expectedTertiary = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x61}
		}
		if got := color.NRGBAModel.Convert(tertiary).(color.NRGBA); got != expectedTertiary {
			t.Errorf("article tertiary color for variant %v = %v, want %v", variant, got, expectedTertiary)
		}
		status := NewTheme().Color(statusMetadataColorName, variant)
		expectedStatus := color.NRGBA{R: 0x78, G: 0x70, B: 0x6A, A: 0xFF}
		if variant == theme.VariantDark {
			expectedStatus = expectedTertiary
		}
		if got := color.NRGBAModel.Convert(status).(color.NRGBA); got != expectedStatus {
			t.Errorf("status metadata color for variant %v = %v, want %v", variant, got, expectedStatus)
		}
	}
}

func TestThemeIconsFollowVariant(t *testing.T) {
	for _, test := range []struct {
		variant fyne.ThemeVariant
		light   fyne.Resource
		dark    fyne.Resource
	}{
		{theme.VariantLight, icEyeLight, icEye},
		{theme.VariantDark, icEyeLight, icEye},
		{theme.VariantLight, icInfoLight, icInfo},
		{theme.VariantDark, icInfoLight, icInfo},
	} {
		want := test.dark
		if test.variant == theme.VariantLight {
			want = test.light
		}
		if got := themeIcon(test.variant, test.light, test.dark).Name(); got != want.Name() {
			t.Errorf("icon for variant %v = %q, want %q", test.variant, got, want.Name())
		}
	}
}

func TestPrimaryActionUsesPeachWithReadableForeground(t *testing.T) {
	for _, variant := range []fyne.ThemeVariant{theme.VariantLight, theme.VariantDark} {
		primary := color.NRGBAModel.Convert(NewTheme().Color(theme.ColorNamePrimary, variant)).(color.NRGBA)
		expectedPeach := color.NRGBA{R: 0xE8, G: 0xAA, B: 0x90, A: 0xFF}
		if variant == theme.VariantDark {
			expectedPeach = color.NRGBA{R: 0xED, G: 0xB2, B: 0x99, A: 0xFF}
		}
		if primary != expectedPeach {
			t.Errorf("primary color for variant %v = %v, want peach", variant, primary)
		}
		foreground := color.NRGBAModel.Convert(NewTheme().Color(theme.ColorNameForegroundOnPrimary, variant)).(color.NRGBA)
		if foreground == primary {
			t.Errorf("primary foreground for variant %v must contrast with the button background", variant)
		}
	}
}

func TestCardHoverTintIsTranslucent(t *testing.T) {
	_, _, _, alpha := cardHoverColor().RGBA()
	if alpha == 0xffff {
		t.Fatal("card hover tint should remain translucent")
	}
}

func TestMetadataSecondaryColorIsMoreProminent(t *testing.T) {
	for _, variant := range []fyne.ThemeVariant{theme.VariantLight, theme.VariantDark} {
		_, _, _, secondaryAlpha := NewTheme().Color(articleMetadataSecondaryColorName, variant).RGBA()
		_, _, _, tertiaryAlpha := NewTheme().Color(articleMetadataTertiaryColorName, variant).RGBA()
		if secondaryAlpha <= tertiaryAlpha {
			t.Errorf("secondary metadata alpha for variant %v = %d, want greater than tertiary %d", variant, secondaryAlpha, tertiaryAlpha)
		}
	}
}

func TestToolbarIconsRenderVisiblePixels(t *testing.T) {
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)

	for name, resource := range map[string]fyne.Resource{
		"refresh":         icRefresh,
		"refresh-rotated": refreshIconFrames()[3],
		"sources":         icLink,
		"power":           icPower,
	} {
		t.Run(name, func(t *testing.T) {
			icon := canvas.NewImageFromResource(resource)
			icon.FillMode = canvas.ImageFillContain
			icon.Resize(fyne.NewSize(32, 32))
			window := fyneTest.NewTempWindow(t, icon)
			window.Resize(fyne.NewSize(32, 32))
			window.Show()

			rendered := window.Canvas().Capture()
			background := color.NRGBAModel.Convert(app.Settings().Theme().Color(theme.ColorNameBackground, app.Settings().ThemeVariant())).(color.NRGBA)
			for y := rendered.Bounds().Min.Y; y < rendered.Bounds().Max.Y; y++ {
				for x := rendered.Bounds().Min.X; x < rendered.Bounds().Max.X; x++ {
					pixel := color.NRGBAModel.Convert(rendered.At(x, y)).(color.NRGBA)
					if pixel != background {
						return
					}
				}
			}
			t.Fatal("icon rendered no visible pixels")
		})
	}
}
