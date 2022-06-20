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
	chart      *Chart
	image      *canvas.Raster
	overlay    *fyne.Container
	background *fyne.Container
}

func newLineChartRenderer(p *Chart) *lineChartRenderer {
	l := &lineChartRenderer{
		chart:      p,
		overlay:    container.NewWithoutLayout(),
		background: container.NewWithoutLayout(),
	}
	l.image = canvas.NewRaster(l.raster)
	return l
}

func (l *lineChartRenderer) AtIndex(dataline, index int) *DataInfo {
	l.chart.locker.Lock()
	defer l.chart.locker.Unlock()

	if index < 0 || index >= len(l.chart.data[dataline]) {
		return nil
	}

	w := l.image.Size().Width
	h := l.image.Size().Height

	zeroY, scale := globalZeroAxisY(
		l.chart,
		fyne.NewSize(w, h),
	)

	v := l.chart.data[dataline][index]
	step := float32(w) / float32(len(l.chart.data[dataline])-1)
	x := int(float32(index) * step)
	y := zeroY - float64(v)*scale
	return &DataInfo{
		Value:    v,
		Position: fyne.NewPos(float32(x), float32(y)),
	}
}

// AtPointer return a data "Point" priving position and value of each point
// closed to the given pos.
func (l *lineChartRenderer) AtPointer(pos fyne.PointEvent) []DataInfo {

	points := make([]DataInfo, len(l.chart.data))
	indexes := make([]int, len(l.chart.data))

	l.chart.locker.Lock()
	for i, d := range l.chart.data {
		step := float32(l.image.Size().Width) / float32(len(d)-1)
		index := int(pos.Position.X / step)
		indexes[i] = index
	}
	l.chart.locker.Unlock()

	for i, index := range indexes {
		points[i] = *l.AtIndex(i, index)
	}
	return points
}

func (l *lineChartRenderer) Destroy() {}

func (l *lineChartRenderer) Layout(size fyne.Size) {
	for _, o := range l.Objects() {
		o.Resize(size)
	}
}

func (l *lineChartRenderer) MinSize() fyne.Size {
	return l.image.MinSize()
}

func (l *lineChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{l.background, l.image, l.overlay}
}

func (l *lineChartRenderer) Overlay() *fyne.Container {
	return l.overlay
}

func (l *lineChartRenderer) Refresh() {
	if l.chart.options.BackgroundColor != nil {
		rect := canvas.NewRectangle(l.chart.options.BackgroundColor)
		rect.Resize(l.image.Size())
		l.background.Objects = []fyne.CanvasObject{rect}
	}
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

	zeroY, scaler := globalZeroAxisY(l.chart, fyne.NewSize(
		float32(scanner.Dest.Bounds().Dx()),
		float32(scanner.Dest.Bounds().Dy()),
	))

	for index, data := range l.chart.data {
		steps := float64(w-int(lineWidth*2)) / float64(len(data)-1)

		lw := float64(lineWidth)
		y := zeroY - float64(data[0])*scaler
		var col color.Color
		if l.chart.options.FillColor != nil {
			fill := createFiller(scanner)
			if len(l.chart.data) > 1 {
				col = l.chart.options.Scheme.ColorAt(index)
				r, g, b, _ := col.RGBA()
				a := 0x99
				col = color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
			} else {
				col = l.chart.options.FillColor
			}
			fill.SetColor(col)
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
