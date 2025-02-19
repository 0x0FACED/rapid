package rapid

import (
	"context"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/0x0FACED/rapid/internal/model"
)

func (a *Rapid) createLANContent() fyne.CanvasObject {
	var addresses []model.ServiceInstance
	addressesCh := make(chan model.ServiceInstance, 10)

	// Запускаем фоновую задачу для поиска серверов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.startDiscovery(ctx, addressesCh)

	// Addresses list
	addressesLabel := widget.NewLabel("Addresses")
	addressesLabel.TextStyle = fyne.TextStyle{Bold: true, Symbol: true}
	addressesLabel.Alignment = fyne.TextAlignCenter
	addressesList := widget.NewList(
		func() int {
			return len(addresses)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Address")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(addresses[i].InstanceName)
		},
	)

	// Обновляем список адресов в UI
	go a.updateAddressList(addressesList, &addresses, addressesCh)

	// Received files list
	receivedFilesLabel := widget.NewLabel("Files from address")
	receivedFilesLabel.TextStyle = fyne.TextStyle{Bold: true, Symbol: true}
	receivedFilesLabel.Alignment = fyne.TextAlignCenter
	receivedFiles := widget.NewList(
		func() int {
			return 0 // Изначально пустой список
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("File")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText("File info" + string(i))
		},
	)

	// Обработка нажатия на адрес
	addressesList.OnSelected = func(id widget.ListItemID) {
		addr := addresses[id].IPv4
		files, err := a.client.GetFiles(addr)
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
			return 0 // Изначально пустой список
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Shared File")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText("Shared file" + string(i))
		},
	)

	// Split the lists
	lists := container.NewHSplit(
		container.NewBorder(
			addressesLabel, // Заголовок сверху
			nil,            // Ничего снизу
			nil,            // Ничего слева
			nil,            // Ничего справа
			addressesList,  // Список адресов в центре (растягивается)
		),
		container.NewVSplit(
			container.NewBorder(
				receivedFilesLabel, // Заголовок сверху
				nil,                // Ничего снизу
				nil,                // Ничего слева
				nil,                // Ничего справа
				receivedFiles,      // Список файлов в центре (растягивается)
			),
			container.NewBorder(
				sharedFilesLabel, // Заголовок сверху
				nil,              // Ничего снизу
				nil,              // Ничего слева
				nil,              // Ничего справа
				sharedFiles,      // Список shared файлов в центре (растягивается)
			),
		),
	)

	return lists
}

func createOptionsContent() fyne.CanvasObject {
	// Placeholder for options content
	return widget.NewLabel("Options Content")
}

func createTopPanel() fyne.CanvasObject {
	// File dialog button
	fileDialogButton := widget.NewButton("Choose file", func() {
		// TODO: Call your API to handle file selection
	})

	// Search bar
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search...")

	// Filters button
	/*filtersButton := widget.NewButton("Filters", func() {
		// TODO: Implement filters functionality
	})*/

	space := layout.NewSpacer()
	space.Resize(fyne.NewSize(10, searchEntry.MinSize().Height))
	// Используем HBox с Spacer для равномерного распределения
	return container.NewGridWithColumns(
		3,
		fileDialogButton,
		searchEntry,
		layout.NewSpacer(),
	)
}

func createFooter() fyne.CanvasObject {
	// Footer with some info
	return widget.NewLabel("Version 1.0.0")
}
