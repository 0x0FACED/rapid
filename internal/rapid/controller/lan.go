package controller

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/0x0FACED/rapid/internal/lan/client"
	"github.com/0x0FACED/rapid/internal/lan/server"
	"github.com/0x0FACED/rapid/internal/model"
)

type LANController struct {
	instName      string
	client        *client.LANClient
	server        *server.LANServer
	serverState   *ServerState
	receivedFiles *FileState
	sharedFiles   *FileState
	serversList   *widget.List
	receivedList  *widget.List
	sharedList    *widget.List
	serversChan   chan model.ServiceInstance
	currentServer string
	refreshTicker *time.Ticker
	shutdownChan  chan struct{}
}

func NewLANController(client *client.LANClient, server *server.LANServer, instName string) *LANController {
	return &LANController{
		instName:      instName,
		client:        client,
		server:        server,
		serverState:   NewServerState(),
		receivedFiles: NewFileState(),
		sharedFiles:   NewFileState(),
		serversChan:   make(chan model.ServiceInstance, 20),
		shutdownChan:  make(chan struct{}),
	}
}

func (lc *LANController) Start(ctx context.Context) {
	go lc.startServerDiscovery(ctx)
	go lc.startServerMaintenance(ctx)
	go lc.processServerUpdates()
}

func (lc *LANController) Stop() {
	close(lc.shutdownChan)
	if lc.refreshTicker != nil {
		lc.refreshTicker.Stop()
	}
	close(lc.serversChan)
}

func (lc *LANController) startServerDiscovery(ctx context.Context) {
	lc.client.DiscoverPeers(ctx, lc.serversChan)
}

func (lc *LANController) startServerMaintenance(ctx context.Context) {
	lc.refreshTicker = time.NewTicker(1 * time.Second)
	defer lc.refreshTicker.Stop()

	for {
		select {
		case <-lc.refreshTicker.C:
			lc.checkServersAvailability()
		case <-lc.shutdownChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (lc *LANController) processServerUpdates() {
	for instance := range lc.serversChan {
		fmt.Println("New Server!")
		isAdded := lc.serverState.AddOrUpdate(instance)
		if isAdded {
			lc.refreshUI()
		}
	}
}

func (lc *LANController) checkServersAvailability() {
	servers := lc.serverState.GetAll()
	for _, server := range servers {
		fmt.Println("Currently pinging: ", server.Address())
		if !lc.client.PingServer(server.Address()) {
			lc.serverState.Remove(server.Key())
			lc.refreshUI()
		}
	}
}

func (lc *LANController) refreshUI() {
	if lc.serversList != nil {
		lc.serversList.Refresh()
	}

	if lc.receivedList != nil {
		lc.receivedList.Refresh()
	}

	if lc.sharedList != nil {
		lc.sharedList.Refresh()
	}
}

func (lc *LANController) CreateLANContent(w fyne.Window) fyne.CanvasObject {
	lc.initServerList()
	lc.initReceivedFilesList()
	lc.initSharedFilesList()

	topPanel := lc.CreateLANTopPanel(w)

	content := container.NewBorder(
		topPanel,
		nil,
		nil,
		nil,
		container.NewHSplit(
			lc.createServerListSection(),
			container.NewVSplit(
				lc.createReceivedFilesSection(),
				lc.createSharedFilesSection(),
			),
		),
	)

	return content
}

func (lc *LANController) initServerList() {
	lc.serversList = widget.NewList(
		func() int { return len(lc.serverState.GetAll()) },
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
			servers := lc.serverState.GetAll()
			if i >= len(servers) {
				return
			}
			server := servers[i]
			container := o.(*fyne.Container)
			labels := container.Objects
			labels[0].(*widget.Label).SetText(server.InstanceName)
			labels[1].(*widget.Label).SetText(server.Address())
		},
	)

	lc.serversList.OnSelected = func(id widget.ListItemID) {
		servers := lc.serverState.GetAll()
		if id >= len(servers) {
			return
		}
		server := servers[id]
		lc.currentServer = server.Key()
		lc.updateReceivedFiles(server)
		lc.serversList.Unselect(id)
	}
	lc.serversList.HideSeparators = true
}

func (lc *LANController) updateReceivedFiles(server model.ServiceInstance) {
	files, err := lc.client.GetFiles(server.IPv4, strconv.Itoa(server.Port))
	if err != nil {
		log.Printf("Error getting files from %s: %v", server.Address(), err)
		return
	}

	lc.receivedFiles = NewFileState()
	for _, file := range files {
		lc.receivedFiles.Add(file.ID, file)
	}
	lc.receivedList.Refresh()
}

func (lc *LANController) initReceivedFilesList() {
	lc.receivedList = widget.NewList(
		func() int { return len(lc.receivedFiles.GetAll()) },
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
			files := lc.receivedFiles.GetAll()
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

	lc.receivedList.OnSelected = func(id widget.ListItemID) {
		files := lc.receivedFiles.GetAll()
		if id >= len(files) {
			return
		}
		file := files[id]
		lc.downloadFile(file)
		lc.receivedList.Unselect(id)
	}

	lc.receivedList.HideSeparators = true
}

func (lc *LANController) downloadFile(file model.File) {
	servers := lc.serverState.GetAll()
	if lc.currentServer == "" || len(servers) == 0 {
		return
	}

	server := lc.findCurrentServer()
	if server == nil {
		return
	}

	err := lc.client.DownloadFile(
		server.IPv4,
		strconv.Itoa(server.Port),
		file.ID,
		file.Name,
	)
	if err != nil {
		log.Printf("Error downloading file %s: %v", file.Name, err)
	}
}

func (lc *LANController) findCurrentServer() *model.ServiceInstance {
	servers := lc.serverState.GetAll()
	for _, server := range servers {
		if server.Key() == lc.currentServer {
			return &server
		}
	}
	return nil
}

func (lc *LANController) initSharedFilesList() {
	lc.sharedList = widget.NewList(
		func() int { return len(lc.sharedFiles.GetAll()) },
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
			files := lc.sharedFiles.GetAll()
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
	lc.sharedList.HideSeparators = true
}

func (lc *LANController) createServerListSection() fyne.CanvasObject {
	spacer := layout.NewSpacer()

	header := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Instance Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		spacer,
		widget.NewLabelWithStyle("Address", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
	)
	label := widget.NewLabelWithStyle("Local Addresses", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	separator := NewCustomSeparator(
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
		2,
		true,
	)

	cont := container.NewBorder(label, header, nil, nil, separator)
	return container.NewBorder(cont, nil, nil, nil, lc.serversList)
}

func (lc *LANController) createReceivedFilesSection() fyne.CanvasObject {
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
		lc.receivedFiles.Filter(query)
		lc.receivedList.Refresh()
	}

	labelCont := container.NewGridWithColumns(2, label, searchEntry)

	cont := container.NewBorder(labelCont, header, nil, nil, separator)
	return container.NewBorder(cont, nil, nil, nil, lc.receivedList)
}

func (lc *LANController) createSharedFilesSection() fyne.CanvasObject {
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
		lc.sharedFiles.Filter(query)
		lc.sharedList.Refresh()
	}

	labelCont := container.NewGridWithColumns(2, label, searchEntry)

	cont := container.NewBorder(labelCont, header, nil, nil, separator)
	return container.NewBorder(cont, nil, nil, nil, lc.sharedList)
}
