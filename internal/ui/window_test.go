package ui

import (
	"slices"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
)

func TestAuxiliaryWindowsUseAppTitle(t *testing.T) {
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)

	showSourcesWindow(app, nil, nil)
	if got := sourcesWin.Title(); got != appName {
		t.Errorf("sources window title = %q, want %q", got, appName)
	}
	sourcesWin.Close()

	showAboutWindow(app)
	if got := aboutWin.Title(); got != appName {
		t.Errorf("about window title = %q, want %q", got, appName)
	}
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
