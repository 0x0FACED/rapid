package rapid

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// TODO: impl
func (a *Rapid) createWebRTCContent() fyne.CanvasObject {
	return widget.NewLabel("test webrtc")
}

// TODO: impl
func createOptionsContent() fyne.CanvasObject {
	return widget.NewLabel("Options Content")
}

func createTopPanel() fyne.CanvasObject {
	fileDialogButton := widget.NewButton("Choose file", func() {
		// TODO: file dialog
	})

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search...")

	space := layout.NewSpacer()
	space.Resize(fyne.NewSize(10, searchEntry.MinSize().Height))
	return container.NewGridWithColumns(
		3,
		fileDialogButton,
		searchEntry,
		layout.NewSpacer(),
	)
}

// TODO: add more widgets
func createFooter() fyne.CanvasObject {
	return widget.NewLabel("Version 1.0.0")
}
