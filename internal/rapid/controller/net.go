package controller

import (
	"bytes"
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/0x0FACED/rapid/internal/lan/server"
	"github.com/0x0FACED/rapid/internal/model"
	"github.com/skip2/go-qrcode"
	"golang.design/x/clipboard"
)

type NetController struct {
	window         *fyne.Window
	instName       string
	receivedFiles  *FileState
	server         *server.LANServer
	sharedFiles    *FileState
	connectionInfo *fyne.Container
	receivedList   *widget.List
	sharedList     *widget.List
	currentServer  string
}

func NewNetController(s *server.LANServer, instName string) *NetController {
	return &NetController{
		instName:      instName,
		server:        s,
		receivedFiles: NewFileState(),
		sharedFiles:   NewFileState(),
	}
}

func (nc *NetController) refreshUI() {
	if nc.receivedList != nil {
		nc.receivedList.Refresh()
	}

	if nc.sharedList != nil {
		nc.sharedList.Refresh()
	}
}

func (nc *NetController) showFilePicker(w fyne.Window) {
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
		if err := nc.handleFileSelection(filePath); err != nil {
			dialog.ShowError(err, w)
		}
	}, w)
}

func (nc *NetController) handleFileSelection(path string) error {
	file, err := nc.server.ShareLocal(path)
	if err != nil {
		return fmt.Errorf("failed to share file: %w", err)
	}

	nc.sharedFiles.Add(file.ID, file)
	nc.sharedList.Refresh()
	return nil
}

func (nc *NetController) CreateLANTopPanel(window fyne.Window) fyne.CanvasObject {
	fileDialogButton := widget.NewButton("Choose File", func() {
		nc.showFilePicker(window)
	})

	name := widget.NewLabelWithStyle("Your name: "+nc.instName, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true, Italic: true})

	cont := container.NewBorder(nil, nil, fileDialogButton, name, nil)
	return cont
}

func (nc *NetController) CreateNetContent(w fyne.Window) fyne.CanvasObject {
	nc.initConnectionInfo(w)
	nc.initReceivedFilesList()
	nc.initSharedFilesList()

	topPanel := nc.CreateLANTopPanel(w)

	content := container.NewBorder(
		topPanel,
		nil,
		nil,
		nil,
		container.NewHSplit(
			nc.createServerListSection(),
			container.NewVSplit(
				nc.createReceivedFilesSection(),
				nc.createSharedFilesSection(),
			),
		),
	)

	return content
}

func (nc *NetController) initConnectionInfo(window fyne.Window) {
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter text for QR code...")

	var qrImage *canvas.Image
	qrImage = canvas.NewImageFromResource(nil)
	qrImage.FillMode = canvas.ImageFillContain
	qrImage.SetMinSize(fyne.NewSquareSize(128))

	clipboard.Init()

	generateBtn := widget.NewButton("Generate QR Code", func() {
		if input.Text == "" {
			dialog.ShowInformation("Error", "Please enter text first", window)
			return
		}

		qr, err := qrcode.New(input.Text, qrcode.Medium)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		var buf bytes.Buffer
		err = qr.Write(128, &buf)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		qrImage.Resource = fyne.NewStaticResource("qr.png", buf.Bytes())
		qrImage.Refresh()

		clipboard.Write(clipboard.FmtImage, qrImage.Resource.Content())
	})

	copyBtn := widget.NewButton("Copy QR Code", func() {
		if qrImage.Resource == nil {
			dialog.ShowInformation("Error", "Generate QR code first", window)
			return
		}

		clipboard.Write(clipboard.FmtImage, qrImage.Resource.Content())
		dialog.ShowInformation("Success", "QR code copied to clipboard", window)
	})

	nc.connectionInfo = container.NewVBox(
		input,
		generateBtn,
		qrImage,
		copyBtn,
	)

}

func (lc *NetController) updateReceivedFiles(server model.ServiceInstance) {
	// TODO: impl
}

func (nc *NetController) initReceivedFilesList() {
	nc.receivedList = widget.NewList(
		func() int { return len(nc.receivedFiles.GetAll()) },
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				widget.NewLabel(""),
				widget.NewLabel(""),
				nil,
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			files := nc.receivedFiles.GetAll()
			if i >= len(files) {
				return
			}
			file := files[i]
			container := o.(*fyne.Container)
			labels := container.Objects
			labels[0].(*widget.Label).SetText(file.Name)
			labels[1].(*widget.Label).SetText(file.SizeString())
		},
	)

	nc.receivedList.OnSelected = func(id widget.ListItemID) {
		files := nc.receivedFiles.GetAll()
		if id >= len(files) {
			return
		}
		file := files[id]
		nc.downloadFile(file)
		nc.receivedList.Unselect(id)
	}

	nc.receivedList.HideSeparators = true
}

func (lc *NetController) downloadFile(file model.File) {
	// TODO: impl
}

func (nc *NetController) initSharedFilesList() {
	nc.sharedList = widget.NewList(
		func() int { return len(nc.sharedFiles.GetAll()) },
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				widget.NewLabel(""),
				widget.NewLabel(""),
				nil,
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			files := nc.sharedFiles.GetAll()
			if i >= len(files) {
				return
			}
			file := files[i]
			container := o.(*fyne.Container)
			labels := container.Objects
			labels[0].(*widget.Label).SetText(file.Name)
			labels[1].(*widget.Label).SetText(file.SizeString())
		},
	)
	nc.sharedList.HideSeparators = true
}

func (nc *NetController) createServerListSection() fyne.CanvasObject {
	header := container.NewGridWithColumns(1,
		widget.NewLabelWithStyle("P2P Connection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	return container.NewBorder(header, nil, nil, nil, nc.connectionInfo)
}

func (nc *NetController) createReceivedFilesSection() fyne.CanvasObject {
	spacer := layout.NewSpacer()

	header := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Filename", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		spacer,
		widget.NewLabelWithStyle("Size", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
	)
	label := widget.NewLabelWithStyle("Received Files", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	separator := NewCustomSeparator(
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
		2,
		true,
	)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search...")
	searchEntry.OnChanged = func(query string) {
		nc.receivedFiles.Filter(query)
		nc.receivedList.Refresh()
	}

	labelCont := container.NewGridWithColumns(2, label, searchEntry)

	cont := container.NewBorder(labelCont, header, nil, nil, separator)
	return container.NewBorder(cont, nil, nil, nil, nc.receivedList)
}

func (nc *NetController) createSharedFilesSection() fyne.CanvasObject {
	spacer := layout.NewSpacer()

	header := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Filename", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		spacer,
		widget.NewLabelWithStyle("Size", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
	)
	label := widget.NewLabelWithStyle("Your Shared files", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	separator := NewCustomSeparator(
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
		2,
		true,
	)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search...")
	searchEntry.OnChanged = func(query string) {
		nc.sharedFiles.Filter(query)
		nc.sharedList.Refresh()
	}

	labelCont := container.NewGridWithColumns(2, label, searchEntry)

	cont := container.NewBorder(labelCont, header, nil, nil, separator)
	return container.NewBorder(cont, nil, nil, nil, nc.sharedList)
}
