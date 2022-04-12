package widget

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"github.com/stretchr/testify/assert"
)

func createGraph() *Graph {
	return NewGraph()
}

func createGraphWithOptions() *Graph {
	return NewGraph(GraphOpts{
		StrokeColor: color.RGBA{0x11, 0x22, 0x33, 255},
		FillColor:   color.RGBA{0x44, 0x55, 0x66, 255},
		StrokeWidth: 5,
		Title: GraphTile{
			Text:  "Test",
			Size:  42,
			Color: color.RGBA{0x77, 0x88, 0x99, 255},
			Style: fyne.TextStyle{Bold: true},
		},
	})
}

func TestCreation(t *testing.T) {
	graph := createGraph()
	if graph == nil {
		t.Error("Creation failed")
	}

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	assert.Equal(t, len(graph.data), 0)
	assert.Equal(t, graph.opts.StrokeColor, theme.ForegroundColor())
	assert.Equal(t, graph.opts.FillColor, theme.DisabledButtonColor())
	assert.Equal(t, graph.opts.StrokeWidth, float32(1))
	assert.Equal(t, graph.opts.Title, GraphTile{
		Text:  "",
		Size:  theme.TextSize(),
		Color: theme.TextColor(),
		Style: fyne.TextStyle{},
	})

	data := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	assert.Equal(t, len(graph.data), len(data))

	// rasterize should be called
	assert.NotEqual(t, graph.image, nil)
}

func TestCreationWithOptions(t *testing.T) {
	graph := createGraphWithOptions()
	if graph == nil {
		t.Error("Creation failed")
	}

	win := test.NewWindow(graph)
	win.Resize(fyne.NewSize(500, 300))
	defer win.Close()

	assert.Equal(t, graph.opts.StrokeColor, color.RGBA{0x11, 0x22, 0x33, 255})
	assert.Equal(t, graph.opts.FillColor, color.RGBA{0x44, 0x55, 0x66, 255})
	assert.Equal(t, graph.opts.StrokeWidth, float32(5))
	assert.Equal(t, graph.opts.Title, GraphTile{
		Text:  "Test",
		Size:  42,
		Color: color.RGBA{0x77, 0x88, 0x99, 255},
		Style: fyne.TextStyle{Bold: true},
	})

	data := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	graph.SetData(data)
	assert.Equal(t, len(graph.data), len(data))

	// rasterize should be called
	assert.NotEqual(t, graph.image, nil)
}
