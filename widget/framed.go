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

// FramedGradient represents a gradient for a framed widget.
type FramedGradient struct {
	ColorSteps FramedGradientStep
	Direction  FramedGradientDirection
}

// FramedOptions represents the options for a framed widget. It's a map with the step position (0 to 1) and a color.
type FramedGradientStep map[float32]color.Color

// FramedGradientDirection represents the direction of a gradient.
type FramedGradientDirection uint8

const (
	// GradientDirectionTopDown represents a gradient from top to bottom.
	GradientDirectionTopDown FramedGradientDirection = iota
	// GradientDirectionBottomUp represents a gradient from bottom to top.
	GradientDirectionLeftRight
)

// FramedOptions holds the options for a framed widget.
type FramedOptions struct {
	// Padding is the amount of padding to add around the content. If °0, the theme.Padding()
	// is used
	Padding float32

	// BorderRadius is the radius of the rounded corners.
	BorderRadius float32

	// BackgroundColor is the color of the background.
	BackgroundColor color.Color

	// BackgroundGradient is the gradient of the background. If set, the BackgroundColor is
	// ignored.
	BackgroundGradient *FramedGradient

	// StrokeColor is the color of the border. Visible only if the
	// StrokeWidth is greater than 0. If not set, the theme.ForegroundColor() is used.
	StrokeColor color.Color

	// StrokeWidth is the width of the border.
	StrokeWidth float32
}

// Framed is the widget modifier to frame a canvas. You can change the border radius, stroke width and color, and use gradients in options.
type Framed struct {
	widget.BaseWidget
	content fyne.CanvasObject
	context *gg.Context
	options *FramedOptions
}

var _ fyne.Widget = (*Framed)(nil)

// NewFramed creates a new framed widget. Options can be nil.
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

	if options.Padding == 0 {
		options.Padding = theme.Padding()
	}

	if options.StrokeColor == nil {
		options.StrokeColor = theme.ForegroundColor()
	}

	// respect scale

	framed.options = options

	framed.ExtendBaseWidget(framed)

	return framed
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
//
// Implements: fyne.Widget
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
		r.framed.options.BorderRadius*r.scaling()+theme.Padding()/2,
		r.framed.options.BorderRadius*r.scaling()+theme.Padding()*2,
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
		if r.framed.options.BackgroundGradient.Direction == GradientDirectionLeftRight {
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
		for i, color := range r.framed.options.BackgroundGradient.ColorSteps {
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
	// gg is inverted in order
	context.FillPreserve()

	if r.framed.options.StrokeWidth > 0 {
		context.SetColor(r.framed.options.StrokeColor)
		context.SetLineWidth(float64(r.framed.options.StrokeWidth * r.scaling()))
		context.Stroke()
	}
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
