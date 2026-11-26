package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type aboutView struct {
	content *fyne.Container
	name    *canvas.Text
	version *canvas.Text
	icon    *widget.Icon
}

func aboutContent(app fyne.App) *aboutView {
	variant := app.Settings().ThemeVariant()
	view := &aboutView{
		name:    canvas.NewText(appName, systemColorForVariant(theme.ColorNameForeground, variant)),
		version: canvas.NewText("versão "+appVersion, systemColorForVariant(theme.ColorNameDisabled, variant)),
		icon:    widget.NewIcon(themeIcon(variant, icEyePeachLight, icEyePeach)),
	}
	view.name.TextSize = 20
	view.name.TextStyle = fyne.TextStyle{Bold: true}
	view.version.TextSize = 12
	description := widget.NewLabelWithStyle(
		"Agregador de notícias direto na sua barra de menu.",
		fyne.TextAlignCenter, fyne.TextStyle{},
	)
	description.Wrapping = fyne.TextWrapWord
	credit := widget.NewLabelWithStyle("© 2026 Clemilson de Azevedo", fyne.TextAlignCenter, fyne.TextStyle{})

	view.content = container.NewPadded(
		container.NewVBox(
			layout.NewSpacer(),
			container.NewCenter(container.NewGridWrap(fyne.NewSize(48, 48), view.icon)),
			container.NewCenter(view.name),
			container.NewCenter(view.version),
			description,
			credit,
			layout.NewSpacer(),
		),
	)
	return view
}

func (v *aboutView) refreshTheme(variant fyne.ThemeVariant) {
	v.name.Color = systemColorForVariant(theme.ColorNameForeground, variant)
	v.name.Refresh()
	v.version.Color = systemColorForVariant(theme.ColorNameDisabled, variant)
	v.version.Refresh()
	v.icon.SetResource(themeIcon(variant, icEyePeachLight, icEyePeach))
}
