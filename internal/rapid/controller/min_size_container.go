package controller

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type MinSizeContainer struct {
	*fyne.Container
	minSize fyne.Size
}

func NewMinSizeContainer(min fyne.Size, obj fyne.CanvasObject) *MinSizeContainer {
	return &MinSizeContainer{
		Container: container.NewWithoutLayout(obj),
		minSize:   min,
	}
}

func (m *MinSizeContainer) MinSize() fyne.Size {
	return m.minSize
}
