package ui

import (
	"image/color"
	"slices"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestAuxiliaryWindowsUseAppTitle(t *testing.T) {
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)
	app.Settings().SetTheme(NewTheme())

	showSourcesWindow(app, nil, nil)
	if got := sourcesWin.Title(); got != appName {
		t.Errorf("sources window title = %q, want %q", got, appName)
	}

	showAboutWindow(app)
	if got := aboutWin.Title(); got != appName {
		t.Errorf("about window title = %q, want %q", got, appName)
	}
	refreshAuxiliaryWindowsTheme(theme.VariantDark)
	wantColor := color.NRGBAModel.Convert(NewTheme().Color(theme.ColorNameForeground, theme.VariantDark)).(color.NRGBA)
	if got := color.NRGBAModel.Convert(sourcesWindowView.heading.Color).(color.NRGBA); got != wantColor {
		t.Errorf("sources heading color = %v, want %v", got, wantColor)
	}
	if got := color.NRGBAModel.Convert(aboutWindowView.name.Color).(color.NRGBA); got != wantColor {
		t.Errorf("about title color = %v, want %v", got, wantColor)
	}
	if got, want := aboutWindowView.icon.Resource.Name(), icEyePeach.Name(); got != want {
		t.Errorf("about window icon = %q, want %q", got, want)
	}

	sourcesWin.Close()
	aboutWin.Close()
}

func TestRemovingSourceUpdatesPreferencesAndNotifies(t *testing.T) {
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)
	app.Preferences().SetStringList(sourcesKey, []string{
		"https://example.com/one.xml",
		"https://example.com/two.xml",
	})

	notifications := 0
	view := newSourcesView(app, nil, func() { notifications++ })
	view.removeSource("https://example.com/one.xml")

	if got, want := loadSources(app), []string{"https://example.com/two.xml"}; !slices.Equal(got, want) {
		t.Errorf("saved sources = %v, want %v", got, want)
	}
	if notifications != 1 {
		t.Errorf("change notifications = %d, want 1", notifications)
	}
}
