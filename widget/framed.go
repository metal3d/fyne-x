package widget

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fogleman/gg"
)

// FramedGradient represents a gradient for a framed widget.
type FramedGradient struct {
	// ColorSteps is a list of colors to use for the gradient. It's a map, the key is the "position" and
	// the value is the color to apply.
	ColorSteps FramedGradientStep

	// Direction to apply the gradient. Defaults to GradientDirectionTopDown. Possibles values are:
	// GradientDirectionTopDown, GradientDirectionLeftRight.
	Direction FramedGradientDirection
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
	// Padding is the amount of padding to add around the content.
	// is used
	Padding float32

	// BorderRadius is the radius of the rounded corners. This apply a padding even if the
	// Padding is set to 0, this to avoid the content to be clipped.
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

// Framed is the widget modifier to frame a canvas. You can change the border radius,
// stroke width and color, and use gradients in options.
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
		options.BackgroundColor = theme.BackgroundColor()
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

// newFramedWidgetRenderer creates the renderer for a framed widget.
func newFramedWidgetRenderer(framed *Framed) fyne.WidgetRenderer {
	renderer := &framedWidgetRenderer{
		framed: framed,
	}
	renderer.img = canvas.NewRaster(renderer.rasterize)
	renderer.container = container.NewMax(renderer.img, framed.content)

	return renderer
}

// Destroy is a private method to Fyne which is called when this widget is deleted.
//
// Implements: fyne.WidgetRenderer
func (r *framedWidgetRenderer) Destroy() {}

// Layout is a private method to Fyne which positions this widget and its children.
//
// Implements: fyne.WidgetRenderer
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

// MinSize is a private method to Fyne which returns the smallest size this widget can shrink to.
//
// Implements: fyne.WidgetRenderer
func (r *framedWidgetRenderer) MinSize() fyne.Size {
	s := r.framed.content.MinSize()

	s = s.Add(fyne.NewSize(
		r.framed.options.BorderRadius*r.scaling(),
		r.framed.options.BorderRadius*r.scaling(),
	))

	return s
}

// Objects is a private method to Fyne which returns this widget's contained objects.
//
// Implements: fyne.WidgetRenderer
func (r *framedWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

// Refresh is a private method to Fyne which updates this widget.
//
// Implements: fyne.WidgetRenderer
func (r *framedWidgetRenderer) Refresh() {
	r.img.Refresh()
}

func (r *framedWidgetRenderer) rasterize(w, h int) image.Image {

	context := gg.NewContext(w, h)

	if r.framed.options.BackgroundGradient != nil {
		var x0, x1, y0, y1 float64
		if r.framed.options.BackgroundGradient.Direction == GradientDirectionLeftRight {
			x1 = float64(w)
		} else {
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
	context.Push()
	context.DrawRoundedRectangle(
		0, 0, float64(w), float64(h),
		float64(r.framed.options.BorderRadius*r.scaling()),
	)
	context.FillPreserve()
	context.Pop()

	if r.framed.options.StrokeWidth > 0 {
		context.Push()
		context.SetColor(r.framed.options.StrokeColor)
		context.SetLineWidth(float64(r.framed.options.StrokeWidth * r.scaling()))
		context.StrokePreserve()
		context.Pop()
	}
	return context.Image()
}

func (r *framedWidgetRenderer) scaling() float32 {
	if fyne.CurrentApp() == nil ||
		fyne.CurrentApp().Driver() == nil ||
		fyne.CurrentApp().Driver().CanvasForObject(r.container) == nil {
		return 1
	}
	return fyne.CurrentApp().Driver().CanvasForObject(r.img).Scale()
}
