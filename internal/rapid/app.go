package rapid

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/0x0FACED/rapid/internal/lan/client"
	"github.com/0x0FACED/rapid/internal/lan/server"
	"github.com/0x0FACED/rapid/internal/rapid/controller"
)

type Rapid struct {
	lan    *server.LANServer
	client *client.LANClient

	lanController *controller.LANController
	netController *controller.NetController

	fyneApp fyne.App

	mu sync.Mutex
}

func New(s *server.LANServer, c *client.LANClient, l *controller.LANController, n *controller.NetController, a fyne.App) *Rapid {
	return &Rapid{
		lan:           s,
		client:        c,
		lanController: l,
		netController: n,
		fyneApp:       a,
	}
}

func (a *Rapid) Start() error {
	main := a.createMainWindow()

	// TODO: refactor
	a.lanController.Start(context.Background())

	main.ShowAndRun()

	return nil
}

func (a *Rapid) createMainWindow() fyne.Window {
	mainWindow := a.fyneApp.NewWindow("rapid")

	tabs := container.NewAppTabs(
		container.NewTabItem("LAN", a.lanController.CreateLANContent(mainWindow)),
		container.NewTabItem("WebRTC", a.netController.CreateNetContent(mainWindow)),
		container.NewTabItem("Options", createOptionsContent()),
	)
	mainWindow.Resize(fyne.NewSize(800, 600))

	content := container.NewBorder(
		nil,
		createFooter(),
		nil,
		nil,
		tabs,
	)

	mainWindow.SetContent(content)
	mainWindow.Resize(fyne.NewSize(800, 600))
	return mainWindow
}
