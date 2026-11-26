package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
)

// ---------------------------------------------------------------- ícones SVG

func svgIcon(name, stroke, body string) fyne.Resource {
	return svgIconWithViewBox(name, stroke, "0 0 24 24", "2", body)
}

func svgIconWithViewBox(name, stroke, viewBox, strokeWidth, body string) fyne.Resource {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="` + viewBox + `" fill="none" stroke="` + stroke +
		`" stroke-width="` + strokeWidth + `" stroke-linecap="round" stroke-linejoin="round">` + body + `</svg>`
	return fyne.NewStaticResource(name+".svg", []byte(svg))
}

func smallSVGIcon(name, stroke, body string) fyne.Resource {
	return svgIconWithViewBox(name, stroke, "-5.14 -5.14 34.28 34.28", "1.5", body)
}

func refreshIconFrames() []fyne.Resource {
	const (
		frameCount = 12
		path       = `<path d="M23 4v6h-6"/><path d="M20.5 15a9 9 0 1 1-2.1-9.4L23 10"/>`
	)
	frames := make([]fyne.Resource, frameCount)
	for frame := range frameCount {
		angle := frame * 360 / frameCount
		body := fmt.Sprintf(`<g transform="rotate(%d 12 12)">%s</g>`, angle, path)
		frames[frame] = svgIconWithViewBox(
			fmt.Sprintf("refresh-%d", angle),
			"#85858D",
			"-5.14 -5.14 34.28 34.28",
			"1.5",
			body,
		)
	}
	return frames
}

var (
	icEye = svgIcon(
		"eye",
		"#F5F5F7",
		`<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>`,
	)
	icEyePeach = svgIcon(
		"eye-peach",
		"#EDB299",
		`<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>`,
	)
	icRefresh = smallSVGIcon("refresh", "#85858D", `<path d="M23 4v6h-6"/><path d="M20.5 15a9 9 0 1 1-2.1-9.4L23 10"/>`)
	icLink    = smallSVGIcon(
		"link",
		"#85858D",
		`<path d="M10 13a5 5 0 0 0 7.5.5l3-3a5 5 0 0 0-7-7l-1.7 1.7"/><path d="M14 11a5 5 0 0 0-7.5-.5l-3 3a5 5 0 0 0 7 7l1.7-1.7"/>`,
	)
	icPower = smallSVGIcon(
		"power",
		"#F2797D",
		`<path d="M18.4 6.6a9 9 0 1 1-12.8 0"/><line x1="12" y1="2" x2="12" y2="12"/>`,
	)
	icDelete = smallSVGIcon(
		"delete",
		"#FF453A",
		`<path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v5"/><path d="M14 11v5"/>`,
	)
	icInfo = svgIcon(
		"info",
		"#A1A1A6",
		`<circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>`,
	)
)
