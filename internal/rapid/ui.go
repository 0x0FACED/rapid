package rapid

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TODO: move

// TODO: impl
func (a *Rapid) createWebRTCContent() fyne.CanvasObject {
	return widget.NewLabel("test webrtc")
}

// TODO: impl
func createOptionsContent() fyne.CanvasObject {
	return widget.NewLabel("Options Content")
}

// TODO: add more widgets
func createFooter() fyne.CanvasObject {
	return widget.NewLabel("Version 1.0.0")
}
