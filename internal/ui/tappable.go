package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// ---------------------------------------------------------------- widget clicável

type tappable struct {
	widget.BaseWidget
	content    fyne.CanvasObject
	background *canvas.Rectangle
	onTap      func()
	hovered    bool
	animation  *fyne.Animation
}

func newTappable(c fyne.CanvasObject, background *canvas.Rectangle, f func()) *tappable {
	t := &tappable{content: c, background: background, onTap: f}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappable) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(t.content) }
func (t *tappable) Tapped(*fyne.PointEvent)             { t.onTap() }
func (t *tappable) Cursor() desktop.Cursor              { return desktop.PointerCursor }
func (t *tappable) MouseIn(*desktop.MouseEvent)         { t.setHovered(true) }
func (t *tappable) MouseMoved(*desktop.MouseEvent)      {}
func (t *tappable) MouseOut()                           { t.setHovered(false) }

func (t *tappable) setHovered(hovered bool) {
	if t.hovered == hovered {
		return
	}
	t.hovered = hovered
	if t.animation != nil {
		t.animation.Stop()
	}

	target := cardColor()
	if hovered {
		target = cardHoverColor()
	}

	t.animation = canvas.NewColorRGBAAnimation(
		t.background.FillColor,
		target,
		140*time.Millisecond,
		func(fill color.Color) {
			t.background.FillColor = fill
			t.background.Refresh()
		},
	)
	t.animation.Start()
}
