package widget

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"sync"
	"text/template"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const svgTplString = `<svg xmlns="http://www.w3.org/2000/svg" width="{{.Width}}" height="{{.Height}}" viewBox="0 0 {{.Width}} {{.Height}}">
    <polygon 
        points="{{range .Data}}{{index . 0}},{{ index . 1}} {{end}}"
        style="fill:{{.Fill}};stroke:{{.StrokeColor}};stroke-width:{{.StrokeWidth}}"
    />
</svg>`

var svgTpl = template.Must(template.New("svg").Parse(svgTplString))

// structure to handle the graph data, colors...
type tplStruct struct {
	Width       int
	Height      int
	Data        [][2]float32
	Fill        string
	StrokeColor string
	StrokeWidth float32
}

// GraphTitle defines a title for the graph.
type GraphTitle struct {
	Color color.Color
	Text  string
	Style fyne.TextStyle
	Size  float32
}

// GraphOpts provides options for the graph.
type GraphOpts struct {

	// FillColor is the color of the fill. Alpha is ignored.
	FillColor color.Color

	// StrokeWidth is the width of the stroke.
	StrokeWidth float32

	// StrokeColor is the color of the stroke. Alpha is ignored.
	StrokeColor color.Color

	// Title is the title of the graph.
	Title GraphTitle
}

// Graph widget provides a plotting widget for data.
type Graph struct {
	widget.BaseWidget
	canvas  *fyne.Container
	overlay *fyne.Container
	data    []float32
	image   *canvas.Raster
	locker  sync.Mutex
	opts    *GraphOpts
	title   *canvas.Text
	yFix    [2]float32

	OnMouseIn    func(*desktop.MouseEvent)
	OnMouseOut   func()
	OnMouseMoved func(*desktop.MouseEvent)
}

// NewGraph creates a new graph widget. The "options" parameter is optional. IF you provide several options, only the first will be used.
func NewGraph(options ...GraphOpts) *Graph {
	if len(options) > 1 {
		log.Println("Warning, too many options passed to NewGraph")
	}
	g := &Graph{
		data:   []float32{},
		locker: sync.Mutex{},
		yFix:   [2]float32{},
	}

	if options != nil {
		g.opts = &options[0]
	} else {
		g.opts = &GraphOpts{
			StrokeWidth: 1,
			StrokeColor: theme.ForegroundColor(),
			FillColor:   theme.DisabledButtonColor(),
			Title:       GraphTitle{},
		}
	}

	if g.opts.StrokeColor == nil {
		g.opts.StrokeColor = theme.ForegroundColor()
	}

	if g.opts.StrokeWidth == 0 {
		g.opts.StrokeWidth = 1
	}

	if g.opts.FillColor == nil {
		g.opts.FillColor = theme.DisabledButtonColor()
	}

	if g.opts.Title.Size == 0 {
		g.opts.Title.Size = theme.TextSize()
	}

	if g.opts.Title.Color == nil {
		g.opts.Title.Color = theme.ForegroundColor()
	}

	g.ExtendBaseWidget(g)

	return g
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (g *Graph) CreateRenderer() fyne.WidgetRenderer {
	g.image = canvas.NewRaster(g.rasterize)
	g.title = canvas.NewText(g.opts.Title.Text, g.opts.Title.Color)
	g.title.TextStyle = g.opts.Title.Style
	g.title.TextSize = g.opts.Title.Size
	g.overlay = container.NewWithoutLayout()
	g.canvas = container.NewWithoutLayout(g.title, g.image, g.overlay)
	return widget.NewSimpleRenderer(g.canvas)
}

// GetDrawable returns the graph's underlying drawable.
func (g *Graph) GetDrawable() *fyne.Container {
	return g.overlay
}

// MinSize returns the smallest size this widget can shrink to.
func (g *Graph) MinSize() fyne.Size {
	if g.image == nil {
		return fyne.NewSize(0, 0)
	}
	return g.image.MinSize()
}

// Refresh refreshes the graph.
func (g *Graph) Refresh() {

	if g.image == nil {
		return
	}
	if g.opts.Title.Text != "" {
		// move the text to the center of the canvas
		g.title.Move(fyne.NewPos(g.image.Size().Width/2-g.title.MinSize().Width/2, 0))
	}
	g.image.Refresh()
	g.canvas.Refresh()
}

// Resize sets a new size for the graph.
func (g *Graph) Resize(size fyne.Size) {
	g.BaseWidget.Resize(size)
	if g.canvas != nil {
		g.image.Resize(size)
		g.canvas.Resize(size)
		g.overlay.Resize(size)
	}
	g.Refresh()
}

// SetData sets the data for the graph - each call to this method will redraw the graph.
func (g *Graph) SetData(data []float32) {
	g.locker.Lock()
	g.data = data
	g.locker.Unlock()
	g.Refresh()
}

// Size returns the size of the graph widget.
func (g *Graph) Size() fyne.Size {
	if g.canvas == nil {
		return fyne.NewSize(0, 0)
	}
	return g.canvas.Size()
}

// This private method is linjed to g.image canvas.Raster property. It uses oksvg and rasterx to render the graph from a SVG template.
func (g *Graph) rasterize(w, h int) image.Image {

	g.locker.Lock()
	defer g.locker.Unlock()

	if g.image == nil || len(g.data) == 0 {
		return image.NewAlpha(image.Rect(0, 0, w, h))
	}

	// <!> Force the width and height to be the same as the image size
	w = int(g.image.Size().Width)
	h = int(g.image.Size().Height)

	// prepare points
	points := make([][2]float32, len(g.data)+2)

	// colors
	fgR, fgG, fgB, _ := g.opts.StrokeColor.RGBA()
	bgR, bgG, bgB, _ := g.opts.FillColor.RGBA()

	// Calculate the max and min values to scale the graph
	// and the step on X to move for each "point"
	width := float32(w)
	height := float32(h)
	stepX := width / float32(len(g.data))
	maxY := float32(0)
	minY := float32(0)
	for _, v := range g.data {
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

	// if we've got a title to draw, reduce the height by the height of the title
	if g.opts.Title.Text != "" {
		maxY += g.opts.Title.Size
	}
	reduce := height / (maxY - minY)

	// keep the Y fix value
	g.yFix = [2]float32{minY, reduce}

	currentX := float32(0)
	points[0] = [2]float32{-g.opts.StrokeWidth, height + minY*reduce + g.opts.StrokeWidth}
	points[len(points)-1] = [2]float32{width + g.opts.StrokeWidth, height + minY*reduce + g.opts.StrokeWidth}

	for i, d := range g.data {
		y := height - d*reduce + minY*reduce
		points[i+1] = [2]float32{currentX, y}
		currentX += stepX
	}

	// render tpl
	buff := new(bytes.Buffer)
	svgTpl.Execute(buff, tplStruct{
		Data:        points,
		Width:       w,
		Height:      h,
		StrokeWidth: g.opts.StrokeWidth,
		StrokeColor: fmt.Sprintf("#%02x%02x%02x", uint8(fgR/0x101), uint8(fgG/0x101), uint8(fgB/0x101)),
		Fill:        fmt.Sprintf("#%02x%02x%02x", uint8(bgR/0x101), uint8(bgG/0x101), uint8(bgB/0x101)),
	})

	// convert the svg to an image
	graph, err := oksvg.ReadIconStream(buff)
	if err != nil {
		log.Println(err)
		return image.NewAlpha(image.Rect(0, 0, w, h))
	}
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	graph.SetTarget(0, 0, float64(w), float64(h))
	scanner := rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())
	graph.Draw(rasterx.NewDasher(w, h, scanner), 1)

	return rgba
}

func (g *Graph) GetDataPosAt(pos fyne.Position) (float32, fyne.Position) {
	stepX := g.image.Size().Width / float32(len(g.data))
	// get the X value corresponding to the data index
	x := int(pos.X / g.image.Size().Width * float32(len(g.data)))
	value := g.data[int(x)]

	// now, get the Y value corresponding to the data value
	y := g.image.Size().Height - value*g.yFix[1] + g.yFix[0]*g.yFix[1]

	// calculate the X value on the graph
	xp := float32(x) * stepX

	return value, fyne.NewPos(xp, y)
}

// implements desktop.Hoverable

func (g *Graph) MouseIn(e *desktop.MouseEvent) {
	if g.OnMouseIn != nil {
		g.OnMouseIn(e)
	}
}

func (g *Graph) MouseMoved(e *desktop.MouseEvent) {
	if g.OnMouseMoved != nil {
		g.OnMouseMoved(e)
	}
}

func (g *Graph) MouseOut() {
	if g.OnMouseOut != nil {
		g.OnMouseOut()
	}
}
