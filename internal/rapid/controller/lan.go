// TODO: impl
package controller

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/0x0FACED/rapid/internal/lan/client"
	"github.com/0x0FACED/rapid/internal/model"
)

// LANController представляет из себя структуру для
// управления контентом списка адресов.
// Здесь
type LANController struct {
	servers       map[string]model.ServiceInstance
	serversSlice  []model.ServiceInstance
	serversCh     chan model.ServiceInstance
	client        *client.LANClient
	mu            sync.Mutex
	serversList   *widget.List
	receivedFiles *widget.List
	sharedFiles   *widget.List
}

func NewLANController(client *client.LANClient) *LANController {
	return &LANController{
		servers:   make(map[string]model.ServiceInstance),
		serversCh: make(chan model.ServiceInstance, 10),
		client:    client,
	}
}
func (c *LANController) OnLength() int {
	return len(c.servers)
}

func (lc *LANController) Start(ctx context.Context) {
	go lc.startScanLocalAddrs(ctx)
	go lc.updateAddressList(ctx)
}

func (lc *LANController) startScanLocalAddrs(ctx context.Context) {
	lc.client.DiscoverServers(ctx, lc.serversCh)
}

func (lc *LANController) updateAddressList(ctx context.Context) {
	go lc.startPingLoop(ctx)

	for addr := range lc.serversCh {
		fmt.Println("NEW ADDR: ", addr.InstanceName)
		if v, exists := lc.servers[addr.IPv4]; !exists || addr.InstanceName != v.InstanceName {
			lc.servers[addr.IPv4+":"+strconv.Itoa(addr.Port)] = addr
			lc.refreshAddressList()
		}
	}
}

func (lc *LANController) startPingLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		lc.mu.Lock()
		for addr, inst := range lc.servers {
			port := strconv.Itoa(inst.Port)
			fmt.Printf("Currently ping: %s:%s inst: %s\n", inst.IPv4, port, inst.InstanceName)
			if !lc.client.PingServer(addr) {
				fmt.Println("deleted server:", inst.InstanceName)
				delete(lc.servers, addr)
			}
		}
		lc.mu.Unlock()
		lc.refreshAddressList()
	}
}

func (lc *LANController) refreshAddressList() {
	lc.serversSlice = make([]model.ServiceInstance, 0, len(lc.servers))
	for _, v := range lc.servers {
		lc.serversSlice = append(lc.serversSlice, v)
	}
	if lc.serversList != nil {
		lc.serversList.Refresh()
	}
}

func (lc *LANController) CreateLANContent() fyne.CanvasObject {
	// Addresses list
	addressesLabel := widget.NewLabel("Addresses")
	addressesLabel.TextStyle = fyne.TextStyle{Bold: true, Symbol: true}
	addressesLabel.Alignment = fyne.TextAlignCenter

	lc.serversList = widget.NewList(
		func() int { return len(lc.serversSlice) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			container := o.(*fyne.Container)
			labels := container.Objects
			labels[0].(*widget.Label).SetText(fmt.Sprintf("%d", i+1))
			labels[1].(*widget.Label).SetText(lc.serversSlice[i].InstanceName)
			labels[2].(*widget.Label).SetText(lc.serversSlice[i].IPv4 + ":" + strconv.Itoa(lc.serversSlice[i].Port))
		},
	)

	lc.serversList.MinSize()

	// Received files list
	receivedFilesLabel := widget.NewLabel("Files from address")
	receivedFilesLabel.TextStyle = fyne.TextStyle{Bold: true, Symbol: true}
	receivedFilesLabel.Alignment = fyne.TextAlignCenter
	lc.receivedFiles = widget.NewList(
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
	lc.serversList.OnSelected = func(id widget.ListItemID) {
		addr := lc.serversSlice[id].IPv4
		port := strconv.Itoa(lc.serversSlice[id].Port)
		files, err := lc.client.GetFiles(addr, port)
		if err != nil {
			log.Println("Error getting files:", err)
			return
		}

		lc.receivedFiles.Length = func() int {
			return len(files)
		}
		lc.receivedFiles.CreateItem = func() fyne.CanvasObject {
			return widget.NewLabel("File")
		}
		lc.receivedFiles.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(files[i].Name)
		}
		lc.receivedFiles.Refresh()
		lc.serversList.Hide()
		lc.serversList.Show()
	}

	// Shared files list
	sharedFilesLabel := widget.NewLabel("Your shared files")
	sharedFilesLabel.TextStyle = fyne.TextStyle{Bold: true, Symbol: true}
	sharedFilesLabel.Alignment = fyne.TextAlignCenter
	lc.sharedFiles = widget.NewList(
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
		container.NewBorder(addressesLabel, nil, nil, nil, lc.serversList),
		container.NewVSplit(
			container.NewBorder(receivedFilesLabel, nil, nil, nil, lc.receivedFiles),
			container.NewBorder(sharedFilesLabel, nil, nil, nil, lc.sharedFiles),
		),
	)

	return lists
}
