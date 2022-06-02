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
	chart *Plot
	image *canvas.Raster
}

func newBarChartRenderer(p *Plot) fyne.WidgetRenderer {
	renderer := &barChartRenderer{
		chart: p,
	}

	renderer.image = canvas.NewRaster(renderer.raster)

	return renderer
}

func (b *barChartRenderer) Destroy() {}

func (b *barChartRenderer) Layout(size fyne.Size) {
	b.image.Resize(size)
}

func (b *barChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(1, 1)
}

func (b *barChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{b.image}
}

func (b *barChartRenderer) Refresh() {
	b.image.Refresh()
}

func (b *barChartRenderer) raster(w, h int) image.Image {
	b.chart.locker.Lock()
	defer b.chart.locker.Unlock()

	if len(b.chart.data) == 0 || len(b.chart.data[0]) == 0 {
		return image.NewAlpha(image.Rect(0, 0, w, h))
	}

	if w == 0 || h == 0 {
		return image.NewAlpha(image.Rect(0, 0, 1, 1))
	}

	lineWidth := b.chart.options.LineWidth * scaling(b.image)
	scanner := createScanner(w, h)
	zeroY, scaler := globalZeroAxisY(b.chart, scanner.Dest)

	for index, data := range b.chart.data {

		steps := float64(w-4) / float64(len(data))

		// create the filler
		fill := createFiller(scanner)
		var col color.Color
		if len(b.chart.data) > 1 {
			col = b.chart.options.Scheme.ColorAt(index)
			red, green, blue, _ := col.RGBA()
			a := 128
			col = color.NRGBA{uint8(red), uint8(green), uint8(blue), uint8(a)}
		} else {
			col = b.chart.options.BackgroundColor
		}
		fill.SetColor(col)

		addLine := float64(lineWidth)

		fill.Start(rasterx.ToFixedP(addLine, zeroY))
		for i, v := range data {
			y := zeroY - float64(v)*scaler
			if y > zeroY {
				y -= addLine
			} else if y < 0 {
				y += addLine
			}
			fill.Line(rasterx.ToFixedP(float64(i)*steps+addLine, y))
			fill.Line(rasterx.ToFixedP(float64(i)*steps+steps, y))
			fill.Line(rasterx.ToFixedP(float64(i)*steps+steps, zeroY))
		}
		fill.Stop(true)
		fill.Draw()

		// create the liner
		line := createStroker(scanner)
		if index > 0 {
			col = b.chart.options.Scheme.ColorAt(index)
			red, green, blue, _ := col.RGBA()
			a := 245
			col = color.NRGBA{uint8(red), uint8(green), uint8(blue), uint8(a)}
		} else {
			col = b.chart.options.LineColor
		}
		line.SetColor(col)
		line.SetStroke(
			fixed.Int26_6(64*lineWidth),
			64,
			nil, nil, nil, rasterx.ArcClip)

		line.Start(rasterx.ToFixedP(float64(lineWidth), zeroY))
		for i, v := range data {
			y := zeroY - float64(v)*scaler
			if y > zeroY {
				y -= addLine
			} else if y < 0 {
				y += addLine
			}
			line.Line(rasterx.ToFixedP(float64(i)*steps+addLine, y))
			line.Line(rasterx.ToFixedP(float64(i)*steps+steps, y))
			line.Line(rasterx.ToFixedP(float64(i)*steps+steps, zeroY))
		}
		line.Stop(false)
		line.Draw()
	}

	return scanner.Dest
}
