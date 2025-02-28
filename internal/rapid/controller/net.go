package controller

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/0x0FACED/rapid/internal/lan/server"
	"github.com/0x0FACED/rapid/internal/model"
	"github.com/caiguanhao/readqr"
	"github.com/skip2/go-qrcode"
	"golang.design/x/clipboard"
)

type NetController struct {
	p2pstate *P2PConnectionState

	window         *fyne.Window
	instName       string
	receivedFiles  *FileState
	server         *server.LANServer
	sharedFiles    *FileState
	connectionInfo *container.Scroll
	receivedList   *widget.List
	sharedList     *widget.List
	currentServer  string
}

func NewNetController(s *server.LANServer, instName string) (*NetController, error) {
	p2pstate, err := NewP2PConnectionState()
	if err != nil {
		return nil, err
	}

	return &NetController{
		p2pstate:      p2pstate,
		instName:      instName,
		server:        s,
		receivedFiles: NewFileState(),
		sharedFiles:   NewFileState(),
	}, nil
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
	clipboard.Init()

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

func (nc *NetController) initCreateConnectionTab(window fyne.Window) fyne.CanvasObject {
	passEntry := widget.NewEntry()
	passEntry.SetPlaceHolder("Enter password...")

	var qrImage *canvas.Image
	qrImage = canvas.NewImageFromResource(nil)
	qrImage.FillMode = canvas.ImageFillContain
	qrImage.SetMinSize(fyne.NewSquareSize(100))

	genQRBtn := widget.NewButton("Create offer QR Code", func() {
		if passEntry.Text == "" {
			dialog.ShowError(errors.New("Provide password for connection"), window)
			return
		}

		nc.p2pstate.SetPassword(passEntry.Text)

		offer, err := nc.p2pstate.CreateEncodedOffer(nil)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		qr, err := qrcode.New(offer, qrcode.Medium)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		var buf bytes.Buffer
		err = qr.Write(400, &buf)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		fmt.Println("PASS: ", passEntry.Text)

		qrImage.Resource = fyne.NewStaticResource("qr.png", buf.Bytes())
		qrImage.Refresh()
	})

	copyQRBtn := widget.NewButton("Copy QR Code", func() {
		if qrImage.Resource == nil {
			dialog.ShowInformation("Error", "Generate QR code first", window)
			return
		}

		_ = clipboard.Write(clipboard.FmtImage, qrImage.Resource.Content())
		dialog.ShowInformation("Success", "QR code copied to clipboard", window)
	})

	pasteQRBtn := widget.NewButton("Paste answer QR Code", func() {
		imgBytes := clipboard.Read(clipboard.FmtImage)
		if imgBytes == nil {
			dialog.ShowInformation("Error", "You must paste QR Code", window)
			return
		}

		result, err := readqr.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		decodedAnswer, err := nc.p2pstate.DecodeAnswer(result)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		if err := nc.p2pstate.ValidatePassword(decodedAnswer.Hash); err != nil {
			dialog.ShowError(fmt.Errorf("invalid password"), window)
			return
		}

		if err := nc.p2pstate.conn.SetRemoteDescription(decodedAnswer.SDP); err != nil {
			dialog.ShowError(err, window)
			return
		}

		go func() {
			if err := nc.p2pstate.WaitForConnection(30 * time.Second); err != nil {
				dialog.ShowError(err, window)
			}
		}()

		nc.p2pstate.onConnect = func() {
			dialog.ShowInformation("Success", "Connected!", window)
		}

		dialog.ShowInformation("Info", "Answer accepted, connecting...", window)
	})

	return container.NewVBox(
		passEntry,
		genQRBtn,
		qrImage,
		copyQRBtn,
		pasteQRBtn,
	)
}

func (nc *NetController) initConnectTab(window fyne.Window) fyne.CanvasObject {
	passEntry := widget.NewEntry()
	passEntry.SetPlaceHolder("Enter password for connect...")

	var qrImage *canvas.Image
	qrImage = canvas.NewImageFromResource(nil)
	qrImage.FillMode = canvas.ImageFillContain
	qrImage.SetMinSize(fyne.NewSquareSize(100))

	genQRBtn := widget.NewButton("Create answer QR Code", func() {
		if passEntry.Text == "" {
			dialog.ShowError(errors.New("Provide password for connection"), window)
			return
		}

		nc.p2pstate.SetPassword(passEntry.Text)

		fmt.Println("PASS SET: ", passEntry.Text)

		answer, err := nc.p2pstate.CreateEncodedAnswer(nil)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		qr, err := qrcode.New(answer, qrcode.Medium)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		var buf bytes.Buffer
		err = qr.Write(400, &buf)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		go func() {
			if err := nc.p2pstate.WaitForConnection(30 * time.Second); err != nil {
				dialog.ShowError(err, window)
			}
		}()

		qrImage.Resource = fyne.NewStaticResource("qr.png", buf.Bytes())
		qrImage.Refresh()
	})

	copyQRBtn := widget.NewButton("Copy QR Code", func() {
		if qrImage.Resource == nil {
			dialog.ShowInformation("Error", "Generate QR code first", window)
			return
		}

		_ = clipboard.Write(clipboard.FmtImage, qrImage.Resource.Content())
		dialog.ShowInformation("Success", "QR code copied to clipboard", window)
	})

	pasteQRBtn := widget.NewButton("Paste offer QR Code", func() {
		nc.p2pstate.SetPassword(passEntry.Text)

		imgBytes := clipboard.Read(clipboard.FmtImage)
		// not image in clipboard
		if imgBytes == nil {
			dialog.ShowInformation("Error", "You must paste QR Code", window)
			return
		}
		buf := bytes.NewBuffer(imgBytes)

		result, err := readqr.Decode(buf)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		// debug output
		fmt.Println(result)

		decodedOffer, err := nc.p2pstate.DecodeOffer(result)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		if err := nc.p2pstate.ValidatePassword(decodedOffer.Hash); err != nil {
			dialog.ShowError(fmt.Errorf("invalid password"), window)
			return
		}

		if err := nc.p2pstate.conn.SetRemoteDescription(decodedOffer.SDP); err != nil {
			dialog.ShowError(err, window)
			return
		}

		if err := nc.p2pstate.AddRemoteICECandidates(); err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Success", "QR code loaded", window)
	})

	return container.NewVBox(
		passEntry,
		pasteQRBtn,
		genQRBtn,
		qrImage,
		copyQRBtn,
	)
}

func (nc *NetController) initConnectionInfo(window fyne.Window) {
	tabs := container.NewAppTabs(
		container.NewTabItem("Host", nc.initCreateConnectionTab(window)),
		container.NewTabItem("Client", nc.initConnectTab(window)),
	)

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter password for connection...")

	nc.connectionInfo = container.NewVScroll(tabs)

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
