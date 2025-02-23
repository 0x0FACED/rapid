package controller

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type CustomSeparator struct {
	widget.BaseWidget
	color      color.Color
	thickness  float32
	horizontal bool
}

func NewCustomSeparator(color color.Color, thickness float32, horizontal bool) *CustomSeparator {
	s := &CustomSeparator{
		color:      color,
		thickness:  thickness,
		horizontal: horizontal,
	}
	s.ExtendBaseWidget(s)
	return s
}

func (s *CustomSeparator) CreateRenderer() fyne.WidgetRenderer {
	line := canvas.NewRectangle(s.color)
	return &separatorRenderer{
		line:       line,
		thickness:  s.thickness,
		horizontal: s.horizontal,
		sep:        s,
	}
}

type separatorRenderer struct {
	line       *canvas.Rectangle
	thickness  float32
	horizontal bool
	sep        *CustomSeparator
}

func (r *separatorRenderer) Layout(size fyne.Size) {
	if r.horizontal {
		r.line.Resize(fyne.NewSize(size.Width, r.thickness))
		r.line.Move(fyne.NewPos(0, (size.Height-r.thickness)/2))
	} else {
		r.line.Resize(fyne.NewSize(r.thickness, size.Height))
		r.line.Move(fyne.NewPos((size.Width-r.thickness)/2, 0))
	}
}

func (r *separatorRenderer) MinSize() fyne.Size {
	if r.horizontal {
		return fyne.NewSize(1, r.thickness)
	}
	return fyne.NewSize(r.thickness, 1)
}

func (r *separatorRenderer) Refresh() {
	r.line.FillColor = r.sep.color
	r.line.Refresh()
}

func (r *separatorRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.line}
}

func (r *separatorRenderer) Destroy() {}
