// TODO: impl
package controller

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/0x0FACED/rapid/internal/model"
)

// AddressesController представляет из себя структуру для
// управления контентом списка адресов.
// Здесь
type AddressesController struct {
	// Храним сервера текущие
	servers map[string]model.ServiceInstance

	mu sync.RWMutex
}

func NewAddressesController() *AddressesController {
	return &AddressesController{
		servers: make(map[string]model.ServiceInstance),
	}
}

func (c *AddressesController) OnLength() int {
	return len(c.servers)
}

func (c *AddressesController) OnCreate() fyne.CanvasObject {
	return widget.NewLabel("Empty")
}

func (c *AddressesController) OnUpdate(i widget.ListItemID, o fyne.CanvasObject) {

}
