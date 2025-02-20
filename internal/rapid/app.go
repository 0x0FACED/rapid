package rapid

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/0x0FACED/rapid/internal/lan/client"
	"github.com/0x0FACED/rapid/internal/lan/server"
	"github.com/0x0FACED/rapid/internal/model"
)

type Rapid struct {
	lan    *server.LANServer
	client *client.LANClient

	fyneApp fyne.App

	mu sync.Mutex
}

func New(s *server.LANServer, c *client.LANClient, a fyne.App) *Rapid {
	return &Rapid{
		lan:     s,
		client:  c,
		fyneApp: a,
	}
}

func (a *Rapid) Start() error {
	main := a.createMainWindow()

	main.ShowAndRun()

	return nil
}

func (a *Rapid) createMainWindow() fyne.Window {
	mainWindow := a.fyneApp.NewWindow("rapid")

	tabs := container.NewAppTabs(
		container.NewTabItem("LAN", a.createLANContent()),
		container.NewTabItem("WebRTC", a.createWebRTCContent()),
		container.NewTabItem("Options", createOptionsContent()),
	)
	mainWindow.Resize(fyne.NewSize(800, 600))

	content := container.NewBorder(
		createTopPanel(),
		createFooter(),
		nil,
		nil,
		tabs,
	)

	mainWindow.SetContent(content)
	mainWindow.Resize(fyne.NewSize(800, 600))
	return mainWindow
}

// TODO: refactor
func (a *Rapid) updateAddressList(list *widget.List, addressList *[]model.ServiceInstance, addresses map[string]model.ServiceInstance, ch chan model.ServiceInstance) {
	go a.startPingLoop(list, addressList, addresses)

	for addr := range ch {
		fmt.Println("NEW ADDR: ", addr.InstanceName)
		if v, exists := addresses[addr.IPv4]; !exists || addr.InstanceName != v.InstanceName {
			addresses[addr.IPv4+":"+strconv.Itoa(addr.Port)] = addr
			a.refreshAddressList(list, addressList, addresses)
		}
	}
}

// TODO: refactor
func (a *Rapid) startPingLoop(list *widget.List, addressList *[]model.ServiceInstance, addresses map[string]model.ServiceInstance) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		for addr, inst := range addresses {
			port := strconv.Itoa(inst.Port)
			fmt.Printf("Currently ping: %s:%s inst: %s\n", inst.IPv4, port, inst.InstanceName)
			if !a.client.PingServer(addr) {
				fmt.Println("deleted server:", inst.InstanceName)
				delete(addresses, addr)
			}
		}
		a.mu.Unlock()
		a.refreshAddressList(list, addressList, addresses)
	}
}

// TODO: refactor
func (a *Rapid) refreshAddressList(list *widget.List, addressList *[]model.ServiceInstance, addresses map[string]model.ServiceInstance) {
	*addressList = make([]model.ServiceInstance, 0, len(addresses))
	for _, v := range addresses {
		*addressList = append(*addressList, v)
	}
	list.Refresh()
}

func (a *Rapid) startScanLocalAddrs(ctx context.Context, ch chan model.ServiceInstance) {
	a.client.DiscoverServers(ctx, ch)
}
