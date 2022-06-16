package charts

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var _ fyne.WidgetRenderer = (*scatterRenderer)(nil) // ensure that the interface is respected
var _ Rasterizer = (*scatterRenderer)(nil)          // ensure that the interface is respected

type scatterRenderer struct {
	chart *Chart
	image *canvas.Raster
}

func newscatterRenderer(c *Chart) *scatterRenderer {
	s := &scatterRenderer{
		chart: c,
	}
	return s
}

// Destroy is for internal use.
//
// Implements: fyne.WidgetRenderer
func (s *scatterRenderer) Destroy() {
}

// Image returns the canvas.Raster that draws the chart.
//
// Implements: Rasterizer
func (s *scatterRenderer) Image() *canvas.Raster {
	return s.image
}

// Layout is a hook that is called if the widget needs to be laid out.
// This should never call Refresh.
//
// Implements: fyne.WidgetRenderer
func (s *scatterRenderer) Layout(_ fyne.Size) {
}

// MinSize returns the minimum size of the widget that is rendered by this renderer.
//
// Implements: fyne.WidgetRenderer
func (s *scatterRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

// Objects returns all objects that should be drawn.
//
// Implements: fyne.WidgetRenderer
func (s *scatterRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{}
}

// Refresh is a hook that is called if the widget has updated and needs to be redrawn.
// This might trigger a Layout.
//
// Implements: fyne.WidgetRenderer
func (s *scatterRenderer) Refresh() {
}
