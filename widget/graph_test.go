package widget

import (
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func createGraph() *Graph {
	return NewGraph(nil)
}

func createGraphWithOptions() *Graph {
	return NewGraph(&GraphOpts{
		StrokeColor: color.RGBA{0x11, 0x22, 0x33, 255},
		FillColor:   color.RGBA{0x44, 0x55, 0x66, 255},
		StrokeWidth: 5,
	})
}

func TestCreation(t *testing.T) {
	graph := createGraph()

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	assert.Equal(t, len(graph.data), 0)
	assert.Equal(t, graph.opts.StrokeColor, theme.ForegroundColor())
	assert.Equal(t, graph.opts.FillColor, theme.DisabledButtonColor())
	assert.Equal(t, graph.opts.StrokeWidth, float32(1))

	data := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	assert.Equal(t, len(graph.data), len(data))

	// rasterize should be called
	assert.NotEqual(t, graph.image, nil)
}

func TestCreationWithOptions(t *testing.T) {
	graph := createGraphWithOptions()

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	assert.Equal(t, graph.opts.StrokeColor, color.RGBA{0x11, 0x22, 0x33, 255})
	assert.Equal(t, graph.opts.FillColor, color.RGBA{0x44, 0x55, 0x66, 255})
	assert.Equal(t, graph.opts.StrokeWidth, float32(5))

	data := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	assert.Equal(t, len(graph.data), len(data))

	// rasterize should be called
	assert.NotEqual(t, graph.image, nil)
}

func TestRasterizer(t *testing.T) {
	graph := createGraph()
	data := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	img := graph.rasterize(200, 400)
	assert.Equal(t, img.Bounds().Size(), image.Point{200, 400})

	graph = createGraphWithOptions()
	graph.SetData(data)
	img = graph.rasterize(200, 400)
	assert.Equal(t, img.Bounds().Size(), image.Point{200, 400})
}

func TestRasterizerWithNegative(t *testing.T) {
	graph := createGraph()
	data := []float32{-1, -2, -3, -4, -5, -6, -7, -8, -9, -10}
	graph.SetData(data)

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	img := graph.rasterize(200, 400)
	assert.Equal(t, img.Bounds().Size(), image.Point{200, 400})

	graph = createGraphWithOptions()
	data = []float32{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4}
	graph.SetData(data)
	img = graph.rasterize(200, 400)
	assert.Equal(t, img.Bounds().Size(), image.Point{200, 400})

}

func TestWithNoData(t *testing.T) {
	graph := createGraph()
	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	assert.Equal(t, len(graph.data), 0)
	assert.Equal(t, graph.opts.StrokeColor, theme.ForegroundColor())
	assert.Equal(t, graph.opts.FillColor, theme.DisabledButtonColor())
	assert.Equal(t, graph.opts.StrokeWidth, float32(1))

	// call rasterizer
	img := graph.rasterize(200, 400)
	assert.Equal(t, img.Bounds().Size(), image.Point{200, 400})
}
