package charts

import (
	"log"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Type defines the kind of plot (pie, bar, line...).
type Type uint8

const (
	// Bar is a bar chart.
	Bar Type = iota
	// Line is a line chart.
	Line
	// Pie is a pie chart.
	Pie
	// Scatter is a scatter chart.
	SCatter
)

var _ fyne.Widget = (*Chart)(nil)
var _ fyne.CanvasObject = (*Chart)(nil)

type ScatterData struct {
	X, Y, Z float32
}

// Chart holds data and internal properties to draw a chart.
type Chart struct {
	widget.BaseWidget
	data     [][]float32
	kind     Type
	options  *Options
	locker   *sync.Mutex
	renderer fyne.WidgetRenderer
}

// NewChart creates a new plot. "kind" defines the kind of plot (pie, bar, line...).
// Options can be nil.
func NewChart(kind Type, options *Options) *Chart {
	plot := &Chart{
		kind:   kind,
		locker: new(sync.Mutex),
	}

	lineWidth := 1.0
	lineColor := theme.PrimaryColor()
	if kind == Pie {
		lineWidth = 0.0
		lineColor = theme.BackgroundColor()
	}
	if options == nil {
		options = &Options{
			FillColor: theme.DisabledButtonColor(),
			LineColor: lineColor,
			LineWidth: float32(lineWidth),
		}
	}

	if options.LineColor == nil {
		options.LineColor = lineColor
	}

	if options.Scheme == nil {
		options.Scheme = AnalogousScheme(nil)
	}

	if options.FillColor == nil && options.LineWidth == 0.0 {
		// indicate a Warning to the user Because the chart can be completely transparent
		log.Println(
			"Warning: BackgroundColor is transparent and lineWidth is 0.0. " +
				"The chart will be completely transparent.",
		)
	}

	plot.options = options
	plot.ExtendBaseWidget(plot)
	return plot
}

// Clear removes all the data from the chart. It will not Refresh the chart view.
func (plot *Chart) Clear() {
	plot.locker.Lock()
	defer plot.locker.Unlock()
	plot.data = [][]float32{}
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
//
// Implements: fyne.Widget
func (plot *Chart) CreateRenderer() fyne.WidgetRenderer {
	switch plot.kind {
	case Bar:
		plot.renderer = newBarChartRenderer(plot)
	case Line:
		plot.renderer = newLineChartRenderer(plot)
	case Pie:
		plot.renderer = newPieChartRenderer(plot)
	default:
		plot.renderer = newLineChartRenderer(plot)
	}
	return plot.renderer
}

// GetYAt returns the value at the given x position.
func (plot *Chart) GetDataAt(pe fyne.PointEvent) []DataInfo {
	if p, ok := plot.renderer.(Pointable); ok {
		return p.AtPointer(pe)
	}
	return nil
}

// Options returns the options of the chart.
func (plot *Chart) Options() *Options {
	return plot.options
}

// Overlay returns the overlay container of the chart.
func (plot *Chart) Overlay() *fyne.Container {
	if o, ok := plot.renderer.(Overlayable); ok {
		return o.Overlay()
	}
	return nil
}

// SetData set all the data for the chart. Because line and bar charts can stack
// several data lines, the data is a 2 dimensional slice.
func (plot *Chart) SetData(data [][]float32) {
	plot.locker.Lock()
	defer plot.locker.Unlock()
	plot.data = data
	plot.Refresh()
}

// Plot adds a new data line to the chart and draw it.
// Warning, PieChart will ignore the added data.
func (plot *Chart) Plot(data []float32) {
	plot.locker.Lock()
	defer plot.locker.Unlock()
	if data != nil && len(data) > 0 {
		plot.data = append(plot.data, data)
	}
	plot.Refresh()
}

// Refresh redraw the chart.
//
// Implements: fyne.Widget
func (plot *Chart) Refresh() {
	if plot.renderer != nil {
		plot.renderer.Refresh()
	}
}

// Resize changes the size of the chart.
//
// Implements: fyne.CanvasObject
func (plot *Chart) Resize(size fyne.Size) {
	if plot.renderer != nil {
		for _, o := range plot.renderer.Objects() {
			o.Resize(size)
		}
	}
}

// Size returns the size of the chart.
//
// Implements: fyne.CanvasObject
func (plot *Chart) Size() fyne.Size {
	if plot.renderer != nil {
		if o, ok := plot.renderer.(Rasterizer); ok {
			return o.Image().Size()
		}
	}
	return fyne.Size{}
}
