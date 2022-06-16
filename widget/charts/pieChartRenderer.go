package charts

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

var _ fyne.WidgetRenderer = (*pieChartRenderer)(nil)
var _ Pointable = (*pieChartRenderer)(nil)
var _ Overlayable = (*pieChartRenderer)(nil)
var _ Rasterizer = (*pieChartRenderer)(nil)

type pieChartRenderer struct {
	chart   *Chart
	image   *canvas.Raster
	overlay *fyne.Container
}

func newPieChartRenderer(chart *Chart) *pieChartRenderer {
	p := &pieChartRenderer{
		chart:   chart,
		overlay: container.NewWithoutLayout(),
	}

	p.image = canvas.NewRaster(p.raster)

	return p
}

func (p *pieChartRenderer) AtPointer(pos fyne.PointEvent) []DataInfo {
	data := p.chart.data[0] // only one series is possible with pie chart
	total := p.sum(data)
	w := p.image.Size().Width
	h := p.image.Size().Height
	center := fyne.NewPos(w/2, h/2)
	r := (math.Min(float64(w), float64(h)) / 2) - float64(p.chart.options.LineWidth)*2
	currAngle := 0.0
	for _, d := range data {
		angle := p.getAngle(total, d)
		y := h - pos.Position.Y
		if p.pointInSector(r, center, fyne.NewPos(pos.Position.X, y), angle, currAngle) {
			point := DataInfo{}
			point.Value = d

			// data coord is in the "inner middle" of the pie slice
			// so we need to calculate the bisector axis and get
			// x and y at the center of the radius
			bisector := currAngle + angle/2
			r /= 2
			px := center.X + float32(r*math.Cos((bisector-90)*math.Pi/180))
			py := center.Y + float32(r*math.Sin((bisector-90)*math.Pi/180))
			point.Position = fyne.NewPos(px, py)

			return []DataInfo{point}
		}
		currAngle += angle
	}
	return []DataInfo{}
}

func (pieChartRenderer) Destroy() {}

func (p *pieChartRenderer) Image() *canvas.Raster {
	return p.image
}

func (p *pieChartRenderer) Layout(size fyne.Size) {
	p.image.Resize(size)
}

func (p *pieChartRenderer) MinSize() fyne.Size {
	return p.image.MinSize()
}

func (p *pieChartRenderer) Refresh() {
	p.image.Refresh()
}

func (p *pieChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{p.image, p.overlay}
}

func (p *pieChartRenderer) Overlay() *fyne.Container {
	return p.overlay
}

func (p *pieChartRenderer) raster(w, h int) image.Image {

	p.chart.locker.Lock()
	defer p.chart.locker.Unlock()

	if len(p.chart.data) == 0 || len(p.chart.data[0]) == 0 {
		return image.NewAlpha(image.Rect(0, 0, w, h))
	}
	if w == 0 || h == 0 {
		return image.NewAlpha(image.Rect(0, 0, 1, 1))
	}

	data := p.chart.data[0]

	scanner := createScanner(w, h)
	scanner.SetColor(color.Transparent)
	var angle float64
	cx, cy := float64(w)/2, float64(h)/2
	r := (math.Min(float64(w), float64(h)) / 2) - float64(p.chart.options.LineWidth)*2
	px, py := cx, cy-r // start on the right side
	total := p.sum(data)

	filler := createFiller(scanner)
	scheme := p.chart.options.Scheme
	for i, d := range data {

		// get random color from Natural
		col := scheme.ColorAt(i)

		filler.SetColor(col)
		angle += p.getAngle(total, d)
		px, py = p.piePart(cx, cy, px, py, r, angle, filler)
	}

	if p.chart.options.LineWidth > 0 {
		stroker := rasterx.NewStroker(w, h, scanner)
		stroker.SetStroke(fixed.Int26_6(64*float64(p.chart.options.LineWidth*scaling(p.image))), 64, nil, nil, nil, rasterx.ArcClip)
		px, py = cx, cy-r
		angle = 0.0
		for _, d := range data {
			// get random color from colorScheme
			stroker.SetColor(p.chart.options.LineColor)
			angle += p.getAngle(total, d)
			px, py = p.piePart(cx, cy, px, py, r, angle, stroker)
		}
	}

	return scanner.Dest
}

func (*pieChartRenderer) sum(data []float32) float64 {
	var sum float32
	for _, d := range data {
		sum += d
	}
	return float64(sum)
}

func (*pieChartRenderer) getAngle(sum float64, d float32) float64 {
	return 360.0 * float64(d) / sum
}

func (*pieChartRenderer) piePart(cx, cy, fx, fy, r, angle float64, ra rasterx.Adder) (float64, float64) {
	rot := (angle - 90) * math.Pi / 180.0
	px := cx + r*math.Cos(rot)
	py := cy + r*math.Sin(rot)

	points := []float64{r, r, 0, 1, 0, fx, fy}
	ra.Start(rasterx.ToFixedP(px, py))
	rasterx.AddArc(points, cx, cy, px, py, ra)
	ra.Line(rasterx.ToFixedP(cx, cy))
	ra.Stop(true)
	ra.(rasterx.Scanner).Draw()
	ra.(rasterx.Scanner).Clear()

	return px, py
}

func (p *pieChartRenderer) pointInSector(radius float64, center, pos fyne.Position, dataAngle, startAngle float64) bool {
	x, y := float64(pos.X-center.X), float64(pos.Y-center.Y)
	endAngle := startAngle + dataAngle
	polarRadius := math.Sqrt(x*x + y*y)

	// angle should be in range [0, 2*pi]
	angle := math.Atan2(y, x) * 360 / (2 * math.Pi)
	angle = -(angle - 90)
	if angle < 0 {
		angle += 360
	}

	return angle >= startAngle && angle <= endAngle && polarRadius < radius
}
