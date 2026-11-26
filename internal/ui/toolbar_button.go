package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type toolbarButton struct {
	widget.BaseWidget
	content       fyne.CanvasObject
	background    *canvas.Circle
	icon          *canvas.Image
	iconResource  fyne.Resource
	iconFrames    []fyne.Resource
	iconAnimation *fyne.Animation
	onTap         func()
	animation     *fyne.Animation
	danger        bool
	enabled       bool
	hovered       bool
	loading       bool
}

func newToolbarButton(icon fyne.Resource, danger bool, onTap func()) *toolbarButton {
	button := &toolbarButton{
		background: canvas.NewCircle(toolbarButtonColor(danger, false)),
		onTap:      onTap,
		danger:     danger,
		enabled:    true,
	}
	button.icon = canvas.NewImageFromResource(icon)
	button.icon.FillMode = canvas.ImageFillContain
	button.iconResource = icon
	iconFrame := container.NewGridWrap(fyne.NewSize(18, 18), button.icon)
	button.content = container.NewGridWrap(
		fyne.NewSize(28, 28),
		container.NewStack(
			button.background,
			container.NewCenter(iconFrame),
		),
	)
	button.ExtendBaseWidget(button)
	return button
}

func (b *toolbarButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.content)
}

func (b *toolbarButton) MinSize() fyne.Size {
	return fyne.NewSize(28, 28)
}

func (b *toolbarButton) Tapped(*fyne.PointEvent) {
	if b.enabled && b.onTap != nil {
		b.onTap()
	}
}

func (b *toolbarButton) Cursor() desktop.Cursor {
	if !b.enabled {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}

func (b *toolbarButton) MouseIn(*desktop.MouseEvent) {
	b.setHovered(true)
}

func (b *toolbarButton) MouseMoved(*desktop.MouseEvent) {}

func (b *toolbarButton) MouseOut() {
	b.setHovered(false)
}

func (b *toolbarButton) Disable() {
	b.enabled = false
	b.setHovered(false)
	b.Refresh()
}

func (b *toolbarButton) Enable() {
	b.enabled = true
	b.Refresh()
}

func (b *toolbarButton) setIconFrames(frames []fyne.Resource) {
	b.iconFrames = frames
}

func (b *toolbarButton) setLoading(loading bool) {
	if b.loading == loading {
		return
	}
	b.loading = loading
	if loading {
		b.Disable()
		if len(b.iconFrames) == 0 {
			return
		}
		b.iconAnimation = &fyne.Animation{
			Duration:    720 * time.Millisecond,
			RepeatCount: fyne.AnimationRepeatForever,
			Curve:       fyne.AnimationLinear,
			Tick: func(progress float32) {
				frame := int(progress*float32(len(b.iconFrames))) % len(b.iconFrames)
				b.icon.Resource = b.iconFrames[frame]
				b.icon.Refresh()
			},
		}
		b.iconAnimation.Start()
		return
	}

	if b.iconAnimation != nil {
		b.iconAnimation.Stop()
		b.iconAnimation = nil
	}
	b.icon.Resource = b.iconResource
	b.icon.Refresh()
	b.Enable()
}

func (b *toolbarButton) setHovered(hovered bool) {
	if (hovered && !b.enabled) || b.hovered == hovered {
		return
	}
	b.hovered = hovered
	if b.animation != nil {
		b.animation.Stop()
	}
	target := toolbarButtonColor(b.danger, hovered)
	b.animation = canvas.NewColorRGBAAnimation(
		b.background.FillColor,
		target,
		140*time.Millisecond,
		func(fill color.Color) {
			b.background.FillColor = fill
			b.background.Refresh()
		},
	)
	b.animation.Start()
}

func toolbarButtonColor(danger, hovered bool) color.NRGBA {
	fill := color.NRGBA{R: 0x78, G: 0x78, B: 0x80, A: 0x1F}
	if danger {
		fill = color.NRGBA{R: 0xFF, G: 0x3B, B: 0x30, A: 0x1F}
	}
	if hovered {
		fill.A = 0x38
	}
	return fill
}
