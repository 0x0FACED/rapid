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

	fyneApp fyne.App

	mu sync.Mutex
}

func New(s *server.LANServer, c *client.LANClient, l *controller.LANController, a fyne.App) *Rapid {
	return &Rapid{
		lan:           s,
		client:        c,
		lanController: l,
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
		container.NewTabItem("LAN", a.lanController.CreateLANContent()),
		container.NewTabItem("WebRTC", a.createWebRTCContent()),
		container.NewTabItem("Options", createOptionsContent()),
	)
	mainWindow.Resize(fyne.NewSize(800, 600))

	content := container.NewBorder(
		a.lanController.CreateTopPanel(mainWindow),
		createFooter(),
		nil,
		nil,
		tabs,
	)

	mainWindow.SetContent(content)
	mainWindow.Resize(fyne.NewSize(800, 600))
	return mainWindow
}
