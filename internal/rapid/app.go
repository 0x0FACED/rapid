package rapid

import (
	"context"

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

func (a Rapid) createMainWindow() fyne.Window {
	mainWindow := a.fyneApp.NewWindow("rapid")

	// 1. Верхняя панель для вкладок
	tabs := container.NewAppTabs(
		container.NewTabItem("LAN", a.createLANContent()),
		container.NewTabItem("WebRTC", a.createLANContent()), // same as lan currenty
		container.NewTabItem("Options", createOptionsContent()),
	)
	mainWindow.Resize(fyne.NewSize(800, 600))

	content := container.NewBorder(
		createTopPanel(), // Top panel with file dialog and search
		createFooter(),   // Footer with additional info
		nil,
		nil,
		tabs,
	)

	mainWindow.SetContent(content)
	mainWindow.Resize(fyne.NewSize(800, 600))
	return mainWindow
}

func (a *Rapid) updateAddressList(list *widget.List, addresses *[]model.ServiceInstance, ch chan model.ServiceInstance) {
	for {
		select {
		case addr := <-ch:
			a.fyneApp.SendNotification(fyne.NewNotification("New server found", addr.InstanceName))
			*addresses = append(*addresses, addr)
			list.Refresh()
		}
	}
}

func (a *Rapid) startDiscovery(ctx context.Context, ch chan model.ServiceInstance) {
	go func() {
		a.client.DiscoverServers(ctx, ch)
	}()
}
