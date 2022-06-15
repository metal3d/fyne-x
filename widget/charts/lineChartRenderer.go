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

var _ fyne.WidgetRenderer = (*lineChartRenderer)(nil)
var _ Pointable = (*lineChartRenderer)(nil)
var _ Rasterizer = (*lineChartRenderer)(nil)
var _ Overlayable = (*lineChartRenderer)(nil)

type lineChartRenderer struct {
	chart   *Chart
	image   *canvas.Raster
	overlay *fyne.Container
}

func newLineChartRenderer(p *Chart) *lineChartRenderer {
	l := &lineChartRenderer{
		chart: p,
	}
	l.overlay = container.NewWithoutLayout()
	l.image = canvas.NewRaster(l.raster)
	return l
}

func (l *lineChartRenderer) AtPointer(pos fyne.PointEvent) []Point {
	l.chart.locker.Lock()
	defer l.chart.locker.Unlock()

	points := make([]Point, len(l.chart.data))
	w := l.image.Size().Width

	for i, d := range l.chart.data {
		step := w / float32(len(d))
		x := int(pos.Position.X / step)
		points[i] = Point{
			Position: fyne.Position{X: float32(x) * step, Y: pos.Position.Y},
			Value:    d[x],
		}
	}

	return points
}

func (l *lineChartRenderer) Destroy() {}

func (l *lineChartRenderer) Layout(size fyne.Size) {
	l.overlay.Resize(size)
	l.image.Resize(size)
}

func (l *lineChartRenderer) MinSize() fyne.Size {
	return l.image.MinSize()
}

func (l *lineChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{l.image, l.overlay}
}

func (l *lineChartRenderer) Overlay() *fyne.Container {
	return l.overlay
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
		if l.chart.options.BackgroundColor != nil {
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
		}

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

func (l *lineChartRenderer) Image() *canvas.Raster {
	return l.image
}
