package charts

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

var _ fyne.WidgetRenderer = (*barChartRenderer)(nil)
var _ Pointable = (*barChartRenderer)(nil)
var _ Rasterizer = (*lineChartRenderer)(nil)
var _ Overlayable = (*barChartRenderer)(nil)

// barChartRenderer renders a bar chart.
type barChartRenderer struct {
	chart   *Chart         // parent plot object that holds data and options
	image   *canvas.Raster // the image of the chart
	overlay *fyne.Container
}

// newBarChartRenderer return the bar chart renderer.
func newBarChartRenderer(p *Chart) fyne.WidgetRenderer {
	renderer := &barChartRenderer{
		chart:   p,
		overlay: container.NewWithoutLayout(),
	}
	renderer.image = canvas.NewRaster(renderer.raster)
	return renderer
}

func (b *barChartRenderer) AtPointer(pos fyne.PointEvent) []Point {
	b.chart.locker.Lock()
	defer b.chart.locker.Unlock()

	points := make([]Point, len(b.chart.data))
	w := b.image.Size().Width
	h := b.image.Size().Height
	zeroY, scale := globalZeroAxisY(
		b.chart,
		image.Rect(0, 0, int(w), int(h)),
	)

	steps := w / float32(largerDataLine(b.chart.data))
	lineWidth := b.chart.options.LineWidth * scaling(b.image)
	barWidth := (steps / float32(len(b.chart.data))) - lineWidth

	for i, d := range b.chart.data {
		x := int(pos.Position.X / steps)
		v := d[x]
		posx := float32(i)*barWidth + float32(x)*steps + barWidth/2
		y := float32(zeroY) - v*float32(scale)
		points[i] = Point{
			Position: fyne.Position{
				X: posx,
				Y: y,
			},
			Value: v,
		}
	}
	return points
}

// Destroy is called when the widget is removed from the GUI.
//
// Implements: fyne.WidgetRenderer
func (b *barChartRenderer) Destroy() {}

// Layout the widget.
//
// Implements: fyne.WidgetRenderer
func (b *barChartRenderer) Layout(size fyne.Size) {
	b.image.Resize(size)
}

// MinSize calculates the minimum size of the widget.
//
// Implements: fyne.WidgetRenderer
func (b *barChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(1, 1)
}

// Objects return the widget content.
//
// Implements: fyne.WidgetRenderer
func (b *barChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{b.image, b.overlay}
}

func (b *barChartRenderer) Overlay() *fyne.Container {
	return b.overlay
}

// Refresh the widget (and redraw the chart).
//
// Implements: fyne.WidgetRenderer
func (b *barChartRenderer) Refresh() {
	b.image.Refresh()
}

// raster generate the chart image.
func (b *barChartRenderer) raster(w, h int) image.Image {
	b.chart.locker.Lock()
	defer b.chart.locker.Unlock()

	if len(b.chart.data) == 0 || len(b.chart.data[0]) == 0 {
		return image.NewAlpha(image.Rect(0, 0, w, h))
	}

	if w == 0 || h == 0 {
		return image.NewAlpha(image.Rect(0, 0, 1, 1))
	}

	var col color.Color // the color of the line or BackgroundColor
	lineWidth := b.chart.options.LineWidth * scaling(b.image)
	steps := (float64(w) - float64(lineWidth)*2) / largerDataLine(b.chart.data)
	barWidth := steps/float64(len(b.chart.data)) - 2*float64(lineWidth)

	// create the rasterizer
	scanner := createScanner(w, h)

	// get the common zeer Y and the reduction factor to make the data
	// to not overflow the rasterizer
	zeroY, scaler := globalZeroAxisY(b.chart, scanner.Dest)

	for index, data := range b.chart.data {
		filler := createFiller(scanner)
		if len(b.chart.data) > 1 {
			col = b.chart.options.Scheme.ColorAt(index)
			red, green, blue, _ := col.RGBA()
			a := 128
			col = color.NRGBA{uint8(red), uint8(green), uint8(blue), uint8(a)}
		} else {
			col = b.chart.options.BackgroundColor
		}
		filler.SetColor(col)
		drawBars(index, lineWidth, barWidth, steps, data, zeroY, scaler, filler)

		liner := createStroker(scanner)
		if index > 0 {
			col = b.chart.options.Scheme.ColorAt(index)
			red, green, blue, _ := col.RGBA()
			a := 245
			col = color.NRGBA{uint8(red), uint8(green), uint8(blue), uint8(a)}
		} else {
			col = b.chart.options.LineColor
		}
		liner.SetColor(col)
		liner.SetStroke(
			fixed.Int26_6(64*lineWidth),
			64,
			nil, nil, nil, rasterx.ArcClip)

		drawBars(index, lineWidth, barWidth, steps, data, zeroY, scaler, liner)
	}

	return scanner.Dest
}

// drawBars draws the bars of the chart for a data index.
func drawBars(index int, lineWidth float32, barWidth, steps float64, data []float32, zeroY float64, scaler float64, drawer rasterx.Scanner) {

	x := float64(index) * barWidth
	drawer.Start(rasterx.ToFixedP(x, zeroY))
	for i, v := range data {
		if i > 0 {
			drawer.Start(rasterx.ToFixedP(x, zeroY))
		}
		y := zeroY - float64(v)*scaler
		if y > zeroY {
			y -= float64(lineWidth)
		} else if y < 0 {
			y += float64(lineWidth)
		}
		drawer.Line(rasterx.ToFixedP(x, y))
		drawer.Line(rasterx.ToFixedP(x+barWidth+float64(lineWidth), y))
		drawer.Line(rasterx.ToFixedP(x+barWidth+float64(lineWidth), zeroY))
		x += steps
	}
	if f, ok := drawer.(*rasterx.Filler); ok {
		f.Stop(false)
	}
	if s, ok := drawer.(*rasterx.Stroker); ok {
		s.Stop(true)
	}
	drawer.Draw()
}
