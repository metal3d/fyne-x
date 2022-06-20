package charts

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// Overlayable is a chart that provides an overlay.
type Overlayable interface {
	// Overlay returns the overlay of the chart.
	Overlay() *fyne.Container
}

// DataInfo represent a point in the drawn chart.
type DataInfo struct {
	// Value is the "pointed" value
	Value float32
	// Position is the position in the chart where the value is drawn.
	Position fyne.Position
}

// Pointable is a chart where a "pointer event" can be used
// to get the data at a given position.
type Pointable interface {
	// AtPointer return the positions and values in
	// the chart at the given pointer position.
	AtPointer(fyne.PointEvent) []DataInfo

	// AtIndex return the positions and values in
	// the chart at the given data index.
	AtIndex(dataline, index int) *DataInfo
}

// Rasterizer represents a chart that uses a canvas.Raster to draw the chart.
type Rasterizer interface {
	// returnthe canvas.Raster used to draw the chart.
	Image() *canvas.Raster
}
