package widget

import (
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fogleman/gg"
)

var framedScaleLock = &sync.Mutex{}

type FramedGradient struct {
	colors    FramedGradientStep
	direction FramedGradientDirection
}

type FramedGradientStep map[float32]color.Color

func NewFramedGradient(colors FramedGradientStep, direction FramedGradientDirection) *FramedGradient {
	return &FramedGradient{
		colors:    colors,
		direction: direction,
	}
}

type FramedGradientDirection uint8

const (
	GradientDirectionTopDown FramedGradientDirection = iota
	GradientDirectionLeftRight
)

type FramedOptions struct {
	Padding            float32
	BorderRadius       float32
	BackgroundColor    color.Color
	BackgroundGradient *FramedGradient
}

type Framed struct {
	widget.BaseWidget
	content fyne.CanvasObject
	context *gg.Context
	options *FramedOptions
}

var _ fyne.Widget = (*Framed)(nil)

func NewFramed(content fyne.CanvasObject, options *FramedOptions) *Framed {
	framed := &Framed{
		content: content,
	}

	if options == nil {
		options = &FramedOptions{}
	}

	if options.BackgroundColor == nil {
		options.BackgroundColor = theme.DisabledButtonColor()
	}

	if options.BorderRadius == 0 {
		options.BorderRadius = theme.Padding() * 2
	}

	if options.Padding == 0 {
		options.Padding = theme.Padding()
	}

	// respect scale

	framed.options = options

	framed.ExtendBaseWidget(framed)

	return framed
}

func (framed *Framed) CreateRenderer() fyne.WidgetRenderer {
	return newFramedWidgetRenderer(framed)
}

// ------

var _ fyne.WidgetRenderer = (*framedWidgetRenderer)(nil)

type framedWidgetRenderer struct {
	img       *canvas.Raster
	framed    *Framed
	container fyne.CanvasObject
}

func newFramedWidgetRenderer(framed *Framed) fyne.WidgetRenderer {
	renderer := &framedWidgetRenderer{
		framed: framed,
	}
	renderer.img = canvas.NewRaster(renderer.rasterize)
	renderer.container = container.NewMax(renderer.img, framed.content)

	return renderer
}

func (r *framedWidgetRenderer) Destroy() {}

func (r *framedWidgetRenderer) Layout(s fyne.Size) {
	r.container.Resize(s)
	r.framed.content.Resize(s.Subtract(fyne.NewSize(
		r.framed.options.BorderRadius*r.scaling(),
		r.framed.options.BorderRadius*r.scaling(),
	)))
	r.framed.content.Move(fyne.NewPos(
		r.framed.options.BorderRadius*r.scaling()/2,
		r.framed.options.BorderRadius*r.scaling()/2,
	))
}

func (r *framedWidgetRenderer) MinSize() fyne.Size {
	s := r.framed.content.MinSize()

	s = s.Add(fyne.NewSize(
		r.framed.options.BorderRadius*r.scaling(),
		r.framed.options.BorderRadius*r.scaling(),
	))

	return s
}

func (r *framedWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

func (r *framedWidgetRenderer) Refresh() {
	r.img.Refresh()
}

func (r *framedWidgetRenderer) rasterize(w, h int) image.Image {

	context := gg.NewContext(w, h)
	if r.framed.options.BackgroundGradient != nil {

		var x0, x1, y0, y1 float64

		if r.framed.options.BackgroundGradient.direction == GradientDirectionLeftRight {
			x0 = 0
			x1 = float64(w)
			y0 = 0
			y1 = 0
		} else {
			x0 = 0
			x1 = 0
			y0 = 0
			y1 = float64(h)
		}

		gradient := gg.NewLinearGradient(x0, y0, x1, y1)
		for i, color := range r.framed.options.BackgroundGradient.colors {
			gradient.AddColorStop(
				float64(i),
				color,
			)
		}
		context.SetFillStyle(gradient)
	} else {
		context.SetColor(r.framed.options.BackgroundColor)
	}

	context.DrawRoundedRectangle(
		0, 0, float64(w), float64(h),
		float64(r.framed.options.BorderRadius*r.scaling()),
	)
	context.Fill()
	return context.Image()
}

func (r *framedWidgetRenderer) scaling() float32 {
	if fyne.CurrentApp() == nil ||
		fyne.CurrentApp().Driver() == nil ||
		fyne.CurrentApp().Driver().CanvasForObject(r.container) == nil {
		return 1
	}
	framedScaleLock.Lock()
	defer framedScaleLock.Unlock()
	return fyne.CurrentApp().Driver().CanvasForObject(r.img).Scale()
}
