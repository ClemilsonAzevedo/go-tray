package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func aboutContent() fyne.CanvasObject {
	name := canvas.NewText(appName, textMain())
	name.TextSize = 20
	name.TextStyle = fyne.TextStyle{Bold: true}
	version := muted("versão "+appVersion, 12, false)
	description := widget.NewLabelWithStyle(
		"Agregador de notícias direto na sua barra de menu.",
		fyne.TextAlignCenter, fyne.TextStyle{},
	)
	description.Wrapping = fyne.TextWrapWord
	credit := widget.NewLabelWithStyle("© 2026 Clemilson de Azevedo", fyne.TextAlignCenter, fyne.TextStyle{})

	return container.NewPadded(
		container.NewVBox(
			layout.NewSpacer(),
			container.NewCenter(container.NewGridWrap(fyne.NewSize(48, 48), widget.NewIcon(icEyePeach))),
			container.NewCenter(name),
			container.NewCenter(version),
			description,
			credit,
			layout.NewSpacer(),
		),
	)
}
