package controller

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (lc *LANController) showFilePicker(w fyne.Window) {
	if w == nil {
		log.Println("Window is nil, cannot show file picker")
		return
	}

	dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if uri == nil {
			return
		}
		defer uri.Close()

		filePath := uri.URI().Path()
		if err := lc.handleFileSelection(filePath); err != nil {
			dialog.ShowError(err, w)
		}
	}, w)
}

func (lc *LANController) handleFileSelection(path string) error {
	file, err := lc.server.ShareLocal(path)
	if err != nil {
		return fmt.Errorf("failed to share file: %w", err)
	}

	lc.sharedFiles.Add(file.ID, file)
	lc.sharedList.Refresh()
	return nil
}

func (lc *LANController) CreateLANTopPanel(window fyne.Window) fyne.CanvasObject {
	fileDialogButton := widget.NewButton("Choose File", func() {
		lc.showFilePicker(window)
	})

	return fileDialogButton
}
