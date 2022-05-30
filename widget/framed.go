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
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
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
	content  fyne.CanvasObject
	context  *gg.Context
	options  *FramedOptions
	renderer fyne.WidgetRenderer
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

	framed.options = options

	framed.ExtendBaseWidget(framed)

	return framed
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
//
// Implements: fyne.Widget
func (framed *Framed) CreateRenderer() fyne.WidgetRenderer {
	framed.renderer = newFramedWidgetRenderer(framed)
	return framed.renderer
}

func (framed *Framed) Options() *FramedOptions {
	if framed.options == nil {
		framed.options = new(FramedOptions)
	}
	return framed.options
}

// SetOption set the frame option. You need to Refresh() to apply the changes.
func (framed *Framed) SetOption(opts FramedOptions) {
	framed.options = &opts
}

func (framed *Framed) Refresh() {
	if framed.content == nil {
		return
	}
	framed.content.Refresh()
	if framed.renderer != nil {
		framed.renderer.Refresh()
	}
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

	sub := (r.framed.options.StrokeWidth + r.framed.options.Padding) * 2
	r.framed.content.Resize(s.Subtract(fyne.NewSize(sub, sub)))

	pad := r.framed.options.Padding + r.framed.options.StrokeWidth
	r.framed.content.Move(fyne.NewPos(
		pad, pad,
	))

}

// MinSize is a private method to Fyne which returns the smallest size this widget can shrink to.
//
// Implements: fyne.WidgetRenderer
func (r *framedWidgetRenderer) MinSize() fyne.Size {
	s := r.framed.content.MinSize()
	pad := (r.framed.options.Padding + r.framed.options.StrokeWidth) * 2
	s = s.Add(fyne.NewSize(
		pad, pad,
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
	r.framed.content.Refresh()
	r.img.Refresh()
}

func (r *framedWidgetRenderer) rasterize(w, h int) image.Image {

	// scale the radius
	radius := float64(r.framed.options.BorderRadius * r.scaling())

	// Create a new context to draw
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())

	// create a filler to fill the background
	filler := rasterx.NewFiller(w, h, scanner)

	if r.framed.options.BackgroundGradient != nil {
		// Direction
		var points [5]float64
		if r.framed.options.BackgroundGradient.Direction == GradientDirectionLeftRight {
			points[2] = 1
		} else {
			points[3] = 1
		}

		// create the gradient
		gradient := rasterx.Gradient{
			Points:   points,
			IsRadial: false,
			Bounds: struct {
				X float64
				Y float64
				W float64
				H float64
			}{
				X: 0,
				Y: 0,
				W: float64(w),
				H: float64(h),
			},
			Matrix: rasterx.Identity,
		}
		// add color stops
		for offset, col := range r.framed.options.BackgroundGradient.ColorSteps {
			_, _, _, alpha := col.RGBA()
			gradient.Stops = append(gradient.Stops, rasterx.GradStop{
				Offset:    float64(offset),
				StopColor: col,
				Opacity:   float64(alpha&0xff) / 255.0,
			})
		}
		filler.SetColor(gradient.GetColorFunction(1))
	} else {
		filler.SetColor(r.framed.options.BackgroundColor)
	}

	// make a rounder corder with radiu
	strokeWidth := float64(r.framed.options.StrokeWidth*r.scaling()) / 2
	rasterx.AddRoundRect(
		strokeWidth, strokeWidth,
		float64(w)-strokeWidth, float64(h)-strokeWidth,
		radius, radius, 0,
		rasterx.QuadraticGap, filler)
	filler.Draw()

	if r.framed.options.StrokeWidth > 0 {
		// create a stroker and stroke the border
		stroker := rasterx.NewStroker(w, h, scanner)
		stroker.SetColor(r.framed.options.StrokeColor)
		linewidth := float64(r.framed.options.StrokeWidth)
		stroker.SetStroke(
			fixed.Int26_6(linewidth*64*float64(r.scaling())),
			fixed.Int26_6(4*64*r.scaling()),
			nil,
			nil,
			nil,
			rasterx.ArcClip)
		rasterx.AddRoundRect(
			linewidth/2, linewidth/2,
			float64(w)-linewidth/2, float64(h)-linewidth/2,
			radius, radius, 0,
			rasterx.QuadraticGap, stroker)
		stroker.Draw()
	}

	return img
}

func (r *framedWidgetRenderer) scaling() float32 {
	if fyne.CurrentApp() == nil ||
		fyne.CurrentApp().Driver() == nil ||
		fyne.CurrentApp().Driver().CanvasForObject(r.container) == nil {
		return 1
	}
	return fyne.CurrentApp().Driver().CanvasForObject(r.img).Scale()
}
