//go:build !darwin

package native

import "fyne.io/fyne/v2"

// ActivateApp is a no-op on platforms that do not need native activation.
func ActivateApp() {}

// MarkTrayIconTemplate is a no-op outside macOS.
func MarkTrayIconTemplate() {}

// StartPopoverPositioner is a no-op outside macOS.
func StartPopoverPositioner(fyne.Window) {}
