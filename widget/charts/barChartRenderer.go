package charts

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

var _ fyne.WidgetRenderer = (*barChartRenderer)(nil)

// barChartRenderer renders a bar chart.
type barChartRenderer struct {
	plot  *Chart         // parent plot object that holds data and options
	image *canvas.Raster // the image of the chart
}

// newBarChartRenderer return the bar chart renderer.
func newBarChartRenderer(p *Chart) fyne.WidgetRenderer {
	renderer := &barChartRenderer{
		plot: p,
	}

	renderer.image = canvas.NewRaster(renderer.raster)

	return renderer
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
	return []fyne.CanvasObject{b.image}
}

// Refresh the widget (and redraw the chart).
//
// Implements: fyne.WidgetRenderer
func (b *barChartRenderer) Refresh() {
	b.image.Refresh()
}

// raster generate the chart image.
func (b *barChartRenderer) raster(w, h int) image.Image {
	b.plot.locker.Lock()
	defer b.plot.locker.Unlock()

	if len(b.plot.data) == 0 || len(b.plot.data[0]) == 0 {
		return image.NewAlpha(image.Rect(0, 0, w, h))
	}

	if w == 0 || h == 0 {
		return image.NewAlpha(image.Rect(0, 0, 1, 1))
	}

	var col color.Color // the color of the line or BackgroundColor
	lineWidth := b.plot.options.LineWidth * scaling(b.image)
	steps := (float64(w) - float64(lineWidth)*2) / largerDataLine(b.plot.data)
	barWidth := steps/float64(len(b.plot.data)) - 2*float64(lineWidth)

	// create the rasterizer
	scanner := createScanner(w, h)

	// get the common zeer Y and the reduction factor to make the data
	// to not overflow the rasterizer
	zeroY, scaler := globalZeroAxisY(b.plot, scanner.Dest)

	for index, data := range b.plot.data {
		filler := createFiller(scanner)
		if len(b.plot.data) > 1 {
			col = b.plot.options.Scheme.ColorAt(index)
			red, green, blue, _ := col.RGBA()
			a := 128
			col = color.NRGBA{uint8(red), uint8(green), uint8(blue), uint8(a)}
		} else {
			col = b.plot.options.BackgroundColor
		}
		filler.SetColor(col)
		drawBars(index, lineWidth, barWidth, steps, data, zeroY, scaler, filler)

		liner := createStroker(scanner)
		if index > 0 {
			col = b.plot.options.Scheme.ColorAt(index)
			red, green, blue, _ := col.RGBA()
			a := 245
			col = color.NRGBA{uint8(red), uint8(green), uint8(blue), uint8(a)}
		} else {
			col = b.plot.options.LineColor
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
