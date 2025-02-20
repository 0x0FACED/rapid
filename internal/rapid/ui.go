package rapid

import (
	"context"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/0x0FACED/rapid/internal/model"
)

// TODO: impl
func (a *Rapid) createWebRTCContent() fyne.CanvasObject {
	return widget.NewLabel("test webrtc")
}

// TODO: refactor
func (a *Rapid) createLANContent() fyne.CanvasObject {
	addresses := make(map[string]model.ServiceInstance)
	var addressList []model.ServiceInstance
	addressesCh := make(chan model.ServiceInstance, 10)

	ctx := context.Background()
	go a.startScanLocalAddrs(ctx, addressesCh)

	// Addresses list
	addressesLabel := widget.NewLabel("Addresses")
	addressesLabel.TextStyle = fyne.TextStyle{Bold: true, Symbol: true}
	addressesLabel.Alignment = fyne.TextAlignCenter

	addressesList := widget.NewList(
		func() int {
			return len(addressList)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Address")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(addressList[i].InstanceName)
		},
	)

	go a.updateAddressList(addressesList, &addressList, addresses, addressesCh)

	// Received files list
	receivedFilesLabel := widget.NewLabel("Files from address")
	receivedFilesLabel.TextStyle = fyne.TextStyle{Bold: true, Symbol: true}
	receivedFilesLabel.Alignment = fyne.TextAlignCenter
	receivedFiles := widget.NewList(
		func() int {
			return 0
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("File")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText("File info " + strconv.Itoa(i))
		},
	)

	// Обработка нажатия на адрес
	addressesList.OnSelected = func(id widget.ListItemID) {
		addr := addressList[id].IPv4
		port := strconv.Itoa(addressList[id].Port)
		files, err := a.client.GetFiles(addr, port)
		if err != nil {
			log.Println("Error getting files:", err)
			return
		}

		receivedFiles.Length = func() int {
			return len(files)
		}
		receivedFiles.CreateItem = func() fyne.CanvasObject {
			return widget.NewLabel("File")
		}
		receivedFiles.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(files[i].Name)
		}
		receivedFiles.Refresh()
	}

	// Shared files list
	sharedFilesLabel := widget.NewLabel("Your shared files")
	sharedFilesLabel.TextStyle = fyne.TextStyle{Bold: true, Symbol: true}
	sharedFilesLabel.Alignment = fyne.TextAlignCenter
	sharedFiles := widget.NewList(
		func() int {
			return 0
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Shared File")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText("Shared file " + strconv.Itoa(i))
		},
	)

	lists := container.NewHSplit(
		container.NewBorder(addressesLabel, nil, nil, nil, addressesList),
		container.NewVSplit(
			container.NewBorder(receivedFilesLabel, nil, nil, nil, receivedFiles),
			container.NewBorder(sharedFilesLabel, nil, nil, nil, sharedFiles),
		),
	)

	return lists
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
