package main

import (
	"fmt"

	"fyne.io/fyne/v2/app"
	"github.com/0x0FACED/rapid/configs"
	"github.com/0x0FACED/rapid/internal/lan/client"
	"github.com/0x0FACED/rapid/internal/lan/mdnss"
	"github.com/0x0FACED/rapid/internal/lan/server"
	"github.com/0x0FACED/rapid/internal/rapid"
	"github.com/0x0FACED/rapid/pkg/generator"
	"github.com/google/uuid"
)

func main() {
	var name string
	var err error
	name, err = generator.GenerateName()
	if err != nil {
		name = uuid.NewString()
	}

	mdnss, err := mdnss.New(name, 8080)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	c := client.New(mdnss)
	_ = c
	s := server.New(configs.LANServerConfig{DownloadsDir: "./test-dir"})
	go s.Start()

	fyneApp := app.NewWithID(name)
	app := rapid.New(s, c, fyneApp)
	app.Start()
}
