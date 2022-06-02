package charts

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

var _ fyne.WidgetRenderer = (*lineChartRenderer)(nil)

type lineChartRenderer struct {
	chart *Plot
	image *canvas.Raster
}

func newLineChartRenderer(p *Plot) fyne.WidgetRenderer {
	l := &lineChartRenderer{
		chart: p,
	}
	l.image = canvas.NewRaster(l.raster)
	return l
}

func (l *lineChartRenderer) Destroy() {}
func (l *lineChartRenderer) Layout(size fyne.Size) {
	l.image.Resize(size)
}
func (l *lineChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(1, 1)
}
func (l *lineChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{l.image}
}
func (l *lineChartRenderer) Refresh() {
	l.image.Refresh()
}

func (l *lineChartRenderer) raster(w, h int) image.Image {
	l.chart.locker.Lock()
	defer l.chart.locker.Unlock()

	if len(l.chart.data) == 0 || len(l.chart.data[0]) == 0 {
		return image.NewAlpha(image.Rect(0, 0, w, h))
	}
	if w == 0 || h == 0 {
		return image.NewAlpha(image.Rect(0, 0, 1, 1))
	}

	lineWidth := l.chart.options.LineWidth * scaling(l.image)

	scanner := createScanner(w, h)
	zeroY, scaler := globalZeroAxisY(l.chart, scanner.Dest)

	for index, data := range l.chart.data {
		steps := float64(w-int(lineWidth*2)) / float64(len(data)-1)
		fill := createFiller(scanner)
		var col color.Color
		if len(l.chart.data) > 1 {
			col = l.chart.options.Scheme.ColorAt(index)
			r, g, b, _ := col.RGBA()
			a := 128
			col = color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
		} else {
			col = l.chart.options.BackgroundColor
		}
		fill.SetColor(col)

		lw := float64(lineWidth)
		y := zeroY - float64(data[0])*scaler
		if y > zeroY {
			y -= lw
		} else if y < zeroY {
			y += lw
		}
		fill.Start(rasterx.ToFixedP(-lw*4, zeroY))
		fill.Line(rasterx.ToFixedP(-lw*4, y))
		for i, v := range data {
			y := zeroY - float64(v)*scaler
			if y > zeroY {
				y -= float64(lineWidth)
			} else if y < zeroY {
				y += float64(lineWidth)
			}
			fill.Line(rasterx.ToFixedP(float64(i)*steps+lw, y))
		}
		// back to zeroY
		fill.Line(rasterx.ToFixedP(float64(w-int(lineWidth*2)), zeroY))
		fill.Stop(true)
		fill.Draw()

		// create the liner
		line := createStroker(scanner)
		if index > 0 {
			col = l.chart.options.Scheme.ColorAt(index)
			r, g, b, _ := col.RGBA()
			a := 245
			col = color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
		} else {
			col = l.chart.options.LineColor
		}
		line.SetColor(col)
		line.SetStroke(
			fixed.Int26_6(64*lineWidth),
			1,
			rasterx.QuadraticCap,
			rasterx.QuadraticCap,
			rasterx.QuadraticGap,
			rasterx.ArcClip)

		y = zeroY - float64(data[0])*scaler
		if y > zeroY {
			y -= lw
		} else if y < zeroY {
			y += lw
		}
		line.Start(rasterx.ToFixedP(-lw*4, zeroY))
		line.Line(rasterx.ToFixedP(-lw*4, y))
		for i, v := range data {
			y = zeroY - float64(v)*scaler
			if y > zeroY {
				y -= lw
			} else if y < zeroY {
				y += lw
			}
			line.Line(rasterx.ToFixedP(float64(i)*steps+lw, y))
		}
		// back to zeroY
		line.Line(rasterx.ToFixedP(float64(w)+lw*2, y))
		line.Stop(false)
		line.Draw()
	}

	return scanner.Dest
}
