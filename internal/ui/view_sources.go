package ui

import (
	"context"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/clemilsonazevedo/look-news/go-tray/internal/news"
)

type sourcesView struct {
	app       fyne.App
	client    news.Service
	onChange  func()
	rows      *fyne.Container
	heading   *canvas.Text
	entry     *widget.Entry
	status    *widget.Label
	addButton *widget.Button
	busy      bool
}

func newSourcesView(app fyne.App, client news.Service, onChange func()) *sourcesView {
	view := &sourcesView{
		app:       app,
		client:    client,
		onChange:  onChange,
		rows:      container.NewVBox(),
		entry:     widget.NewEntry(),
		status:    widget.NewLabel(""),
		addButton: widget.NewButton("Adicionar", nil),
	}
	view.entry.SetPlaceHolder("https://site.com/feed.xml")
	view.status.Hide()
	view.addButton.Importance = widget.HighImportance
	view.addButton.Disable()
	view.addButton.OnTapped = view.addSource
	view.entry.OnSubmitted = func(string) { view.addSource() }
	view.entry.OnChanged = view.handleEntryChange
	return view
}

func (v *sourcesView) content() fyne.CanvasObject {
	v.heading = canvas.NewText("Fontes RSS", textMain())
	v.heading.TextSize = 14
	v.heading.TextStyle = fyne.TextStyle{Bold: true}
	form := container.New(layout.NewCustomPaddedVBoxLayout(8),
		v.heading,
		container.NewBorder(nil, nil, nil, v.addButton, v.entry),
	)
	top := container.New(layout.NewCustomPaddedLayout(8, 0, 12, 12),
		container.New(layout.NewCustomPaddedVBoxLayout(0), form, v.status),
	)
	sourceList := container.New(layout.NewCustomPaddedLayout(0, 8, 12, 12), container.NewVScroll(v.rows))
	return container.NewBorder(top, nil, nil, nil, sourceList)
}

func (v *sourcesView) refreshTheme(variant fyne.ThemeVariant) {
	v.heading.Color = systemColorForVariant(theme.ColorNameForeground, variant)
	v.heading.Refresh()
	v.entry.Refresh()
	v.status.Refresh()
	v.addButton.Refresh()
	v.rows.Refresh()
}

func (v *sourcesView) rebuildSources() {
	v.rows.Objects = nil
	sources := loadSources(v.app)
	if len(sources) == 0 {
		v.rows.Add(widget.NewLabel("Nenhuma fonte cadastrada."))
	}
	for _, source := range sources {
		v.rows.Add(v.sourceRow(source))
	}
	v.rows.Refresh()
}

func (v *sourcesView) sourceRow(source string) fyne.CanvasObject {
	label := widget.NewLabel(source)
	label.Truncation = fyne.TextTruncateEllipsis
	removeButton := newToolbarButton(icDelete, true, func() { v.removeSource(source) })
	return container.NewVBox(
		container.NewBorder(nil, nil, nil, removeButton, label),
		widget.NewSeparator(),
	)
}

func (v *sourcesView) removeSource(source string) {
	sources := loadSources(v.app)
	index := slices.Index(sources, source)
	if index < 0 {
		return
	}

	sources = slices.Delete(sources, index, index+1)
	v.app.Preferences().SetStringList(sourcesKey, sources)
	v.rebuildSources()
	v.notifyChange()
}

func (v *sourcesView) addSource() {
	if v.busy {
		return
	}

	source, err := news.NormalizeSourceURL(v.entry.Text)
	if err != nil {
		v.setStatus(err.Error())
		return
	}
	if slices.Contains(loadSources(v.app), source) {
		v.setStatus("Essa fonte já foi adicionada.")
		return
	}

	v.entry.SetText(source)
	v.setBusy(true)
	v.setStatus("")
	go func() {
		err := v.client.ValidateFeed(context.Background(), source)
		fyne.Do(func() { v.finishAddingSource(source, err) })
	}()
}

func (v *sourcesView) finishAddingSource(source string, err error) {
	v.setBusy(false)
	if err != nil {
		v.setStatus(err.Error())
		return
	}

	sources := loadSources(v.app)
	if slices.Contains(sources, source) {
		v.setStatus("Essa fonte já foi adicionada.")
		return
	}
	v.app.Preferences().SetStringList(sourcesKey, append(sources, source))
	v.entry.SetText("")
	v.setStatus("")
	v.rebuildSources()
	v.notifyChange()
}

func (v *sourcesView) handleEntryChange(value string) {
	v.setStatus("")
	if v.busy || strings.TrimSpace(value) == "" {
		v.addButton.Disable()
		return
	}
	v.addButton.Enable()
}

func (v *sourcesView) setBusy(busy bool) {
	v.busy = busy
	if busy {
		v.entry.Disable()
		v.addButton.Disable()
		v.addButton.SetText("Verificando…")
		return
	}

	v.entry.Enable()
	v.addButton.SetText("Adicionar")
	v.handleEntryChange(v.entry.Text)
}

func (v *sourcesView) setStatus(message string) {
	v.status.SetText(message)
	if message == "" {
		v.status.Hide()
		return
	}
	v.status.Show()
}

func (v *sourcesView) notifyChange() {
	if v.onChange != nil {
		v.onChange()
	}
}
