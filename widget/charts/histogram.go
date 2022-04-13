package charts

import (
	"bytes"
	"fmt"
	"image"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

type HistrogramChart struct {
	*LineChart // actually it's the same, we only need to change the rasterize method
}

type BarChart = HistrogramChart
type HistChartOptions = LineCharthOpts
type BarchartOptions = HistChartOptions

func NewHistrogramChart(opts *HistChartOptions) *HistrogramChart {
	hist := new(HistrogramChart)
	hist.LineChart = NewLineChart(opts)
	return hist
}

func (hist *HistrogramChart) CreateRenderer() fyne.WidgetRenderer {
	w := hist.LineChart.CreateRenderer()
	hist.LineChart.image = canvas.NewRaster(hist.rasterize)
	hist.LineChart.canvas = container.NewWithoutLayout(hist.LineChart.image, hist.LineChart.overlay)
	w.Objects()[0] = hist.LineChart.canvas

	return w
}

func (hist *HistrogramChart) rasterize(w, h int) image.Image {
	if hist.canvas == nil || len(hist.data) == 0 {
		return image.NewAlpha(image.Rect(0, 0, w, h))
	}

	// <!> Force the width and height to be the same as the image size
	// To not do this will cause the graph to be scaled down.
	// TODO: why is this needed?
	w = int(hist.image.Size().Width)
	h = int(hist.image.Size().Height)

	// Calculate the max and min values to scale the graph
	// and the step on X to move for each "point"
	width := float64(w)
	height := float64(h)
	stepX := width / float64(len(hist.data))
	maxY := float64(0)
	minY := float64(0)
	for _, v := range hist.data {
		if v > maxY {
			maxY = v
		}
		if v < minY {
			minY = v
		}
	}

	// Move the graph to avoid the "zero" line
	if minY > 0 {
		minY = 0
	}

	// reduction factor
	reduce := height / (maxY - minY)

	// keep the Y fix value - used by GetDataPosAt()
	hist.yFix = [2]float64{minY, reduce}

	// Draw...
	currentX := float64(0)

	// each "value" has 4 points (bottom left, top left, top right, bottom right)
	// each point is defined by 2 coordinates (x, y)
	points := make([][2]float64, len(hist.data)*4+1)

	sw := float64(hist.opts.StrokeWidth)

	for i, v := range hist.data {
		// Calculate the points
		// bottom left
		points[i*4+0][0] = currentX
		points[i*4+0][1] = height + sw
		// top left
		points[i*4+1][0] = currentX
		points[i*4+1][1] = height - (v-minY)*reduce + sw
		// top right
		points[i*4+2][0] = currentX + stepX
		points[i*4+2][1] = height - (v-minY)*reduce + sw
		// bottom right
		points[i*4+3][0] = currentX + stepX
		points[i*4+3][1] = height + sw

		currentX += stepX
	}

	points[len(points)-1][0] = currentX
	points[len(points)-1][1] = height

	// colors
	fgR, fgG, fgB, _ := hist.opts.StrokeColor.RGBA()
	bgR, bgG, bgB, _ := hist.opts.FillColor.RGBA()
	// convert the svg to an image.Image
	buff := new(bytes.Buffer)
	svgTpl.Execute(buff, tplStruct{
		Data:        points,
		Width:       w,
		Height:      h,
		StrokeWidth: hist.opts.StrokeWidth,
		StrokeColor: fmt.Sprintf("#%02x%02x%02x", uint8(fgR/0x101), uint8(fgG/0x101), uint8(fgB/0x101)),
		Fill:        fmt.Sprintf("#%02x%02x%02x", uint8(bgR/0x101), uint8(bgG/0x101), uint8(bgB/0x101)),
	})

	graph, err := oksvg.ReadIconStream(buff)
	if err != nil {
		log.Println(err)
		return image.NewRGBA(image.Rect(0, 0, w, h))
	}
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	graph.SetTarget(0, 0, float64(w), float64(h))
	scanner := rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())
	graph.Draw(rasterx.NewDasher(w, h, scanner), 1)
	return rgba

}
