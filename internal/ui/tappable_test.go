package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fyneTest "fyne.io/fyne/v2/test"
)

func TestTappableHoverAndClick(t *testing.T) {
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)

	background := canvas.NewRectangle(cardColor())
	tapped := false
	item := newTappable(container.NewStack(background), background, func() { tapped = true })

	item.MouseIn(nil)
	if !item.hovered {
		t.Fatal("hovering should activate the visual state")
	}
	item.animation.Stop()

	item.Tapped(nil)
	if !tapped {
		t.Fatal("tapping should invoke the article action")
	}

	item.MouseOut()
	if item.hovered {
		t.Fatal("leaving the item should deactivate the visual state")
	}
	item.animation.Stop()
}

func TestToolbarButtonCentersIconAndAnimatesRefresh(t *testing.T) {
	app := fyneTest.NewApp()
	t.Cleanup(app.Quit)

	button := newToolbarButton(icRefresh, false, nil)
	button.setIconFrames(refreshIconFrames())
	button.content.Resize(fyne.NewSize(28, 28))
	stack := button.content.(*fyne.Container).Objects[0].(*fyne.Container)
	center := stack.Objects[1].(*fyne.Container)
	iconFrame := center.Objects[0]
	if got, want := iconFrame.Position(), fyne.NewPos(5, 5); got != want {
		t.Errorf("icon position = %v, want centered at %v", got, want)
	}
	if got, want := iconFrame.Size(), fyne.NewSize(18, 18); got != want {
		t.Errorf("icon size = %v, want %v", got, want)
	}

	button.setLoading(true)
	if button.enabled || !button.loading || button.iconAnimation == nil {
		t.Fatal("loading should disable the button and start its icon animation")
	}
	button.iconAnimation.Stop()
	button.iconAnimation.Tick(0.5)
	if got, want := button.icon.Resource.Name(), button.iconFrames[6].Name(); got != want {
		t.Errorf("animated icon = %q, want frame %q", got, want)
	}

	button.setLoading(false)
	if !button.enabled || button.loading || button.iconAnimation != nil {
		t.Fatal("finishing loading should enable the button and stop its animation")
	}
	if got, want := button.icon.Resource.Name(), icRefresh.Name(); got != want {
		t.Errorf("restored icon = %q, want %q", got, want)
	}
}
