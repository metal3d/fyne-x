package charts

import (
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func createGraph() *LineChart {
	return NewLineChart(nil)
}

func createGraphWithOptions() *LineChart {
	return NewLineChart(&LineCharthOpts{
		StrokeColor: color.RGBA{0x11, 0x22, 0x33, 255},
		FillColor:   color.RGBA{0x44, 0x55, 0x66, 255},
		StrokeWidth: 5,
	})
}

func makeRasterize(win fyne.Window, graph *LineChart) image.Image {
	win.Resize(fyne.NewSize(500, 300))
	img := graph.rasterize(int(graph.Size().Width), int(graph.Size().Height))
	return img
}

func assertSize(t *testing.T, img image.Image, graph *LineChart) {
	assert.Greater(t, img.Bounds().Size().X, 0)
	assert.Greater(t, img.Bounds().Size().Y, 0)
	assert.Equal(t, img.Bounds().Size().X, int(graph.Size().Width))
	assert.Equal(t, img.Bounds().Size().Y, int(graph.Size().Height))
}

func TestGraph_Creation(t *testing.T) {
	graph := createGraph()

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	assert.Equal(t, len(graph.data), 0)
	assert.Equal(t, graph.opts.StrokeColor, theme.ForegroundColor())
	assert.Equal(t, graph.opts.FillColor, theme.DisabledButtonColor())
	assert.Equal(t, graph.opts.StrokeWidth, float32(1))

	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	assert.Equal(t, len(graph.data), len(data))

	// rasterize should be called
	assert.NotEqual(t, graph.image, nil)
}

func TestGraph_CreationWithOptions(t *testing.T) {
	graph := createGraphWithOptions()

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	assert.Equal(t, graph.opts.StrokeColor, color.RGBA{0x11, 0x22, 0x33, 255})
	assert.Equal(t, graph.opts.FillColor, color.RGBA{0x44, 0x55, 0x66, 255})
	assert.Equal(t, graph.opts.StrokeWidth, float32(5))

	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	assert.Equal(t, len(graph.data), len(data))

	// rasterize should be called
	assert.NotEqual(t, graph.image, nil)
}

func TestGraph_Rasterizer(t *testing.T) {
	graph := createGraph()
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(50, 70))
	defer win.Close()

	graph = createGraphWithOptions()
	graph.Resize(fyne.NewSize(50, 70))
	graph.SetData(data)
	img := makeRasterize(win, graph)

	assertSize(t, img, graph)

}

func TestGraph_RasterizerWithNegative(t *testing.T) {
	graph := createGraph()
	data := []float64{-1, -2, -3, -4, -5, -6, -7, -8, -9, -10}
	graph.SetData(data)

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	graph.Resize(fyne.NewSize(500, 300))

	img := makeRasterize(win, graph)
	assertSize(t, img, graph)

	graph = createGraphWithOptions()
	data = []float64{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4}
	graph.SetData(data)
	graph.Resize(fyne.NewSize(500, 300))
	img = makeRasterize(win, graph)
	assertSize(t, img, graph)
}

func TestGraph_WithNoData(t *testing.T) {
	graph := createGraph()
	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	assert.Equal(t, len(graph.data), 0)
	assert.Equal(t, graph.opts.StrokeColor, theme.ForegroundColor())
	assert.Equal(t, graph.opts.FillColor, theme.DisabledButtonColor())
	assert.Equal(t, graph.opts.StrokeWidth, float32(1))

	// call rasterizer
	img := makeRasterize(win, graph)
	assertSize(t, img, graph)
}

func TestGraph_GetOpts(t *testing.T) {
	opts := &LineCharthOpts{
		StrokeColor: color.RGBA{0x11, 0x22, 0x33, 255},
		FillColor:   color.RGBA{0x44, 0x55, 0x66, 255},
		StrokeWidth: 5,
	}
	graph := NewLineChart(opts)

	assert.Equal(t, graph.opts, opts)
	// in case of, check all fields
	assert.Equal(t, graph.opts.StrokeColor, color.RGBA{0x11, 0x22, 0x33, 255})
	assert.Equal(t, graph.opts.FillColor, color.RGBA{0x44, 0x55, 0x66, 255})
	assert.Equal(t, graph.opts.StrokeWidth, float32(5))
}

func TestGraph_GetValAndCurvePos(t *testing.T) {
	graph := createGraph()
	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	graph.CreateRenderer()
	graph.Resize(fyne.NewSize(500, 300))

	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	graph.rasterize(500, 300)

	// Get the value at the center of the graph
	x, y := graph.GetDataPosAt(fyne.NewPos(289, 200))
	assert.Equal(t, float64(6), x)
	assert.Equal(t, float32(250), y.X)
	assert.Equal(t, float32(120), y.Y)
}

func TestGraph_Mouse(t *testing.T) {
	control := 0

	graph := createGraph()
	graph.OnMouseIn = func(e *desktop.MouseEvent) {
		control++
	}
	graph.OnMouseOut = func() {
		control++
	}
	graph.OnMouseMoved = func(e *desktop.MouseEvent) {
		control++
	}
	graph.OnTapped = func(e *fyne.PointEvent) {
		control++
	}

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	graph.CreateRenderer()
	graph.Resize(fyne.NewSize(500, 300))

	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	graph.rasterize(500, 300)

	// trigger all mouse events
	graph.MouseIn(&desktop.MouseEvent{})
	graph.MouseOut()
	graph.MouseMoved(&desktop.MouseEvent{})
	graph.Tapped(&fyne.PointEvent{})

	assert.Equal(t, control, 4)

}
