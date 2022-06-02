package charts

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*Plot)(nil)
var _ Plotable = (*Plot)(nil)

type PlotType uint8

const (
	PlotBar PlotType = iota
	PlotLine
	PlotPie
)

type PlotOptions struct {
	LineWidth       float32
	BackgroundColor color.Color
	LineColor       color.Color
	Scheme          Scheme
}

type Plot struct {
	widget.BaseWidget
	data    [][]float32
	kind    PlotType
	options *PlotOptions
	locker  *sync.Mutex
}

func NewPlot(kind PlotType, options *PlotOptions) *Plot {
	plot := &Plot{
		kind:   kind,
		locker: new(sync.Mutex),
	}

	lineWidth := 1.0
	lineColor := theme.PrimaryColor()
	if kind == PlotPie {
		lineWidth = 0.0
		lineColor = theme.BackgroundColor()
	}
	if options == nil {
		options = &PlotOptions{
			BackgroundColor: theme.DisabledButtonColor(),
			LineColor:       lineColor,
			LineWidth:       float32(lineWidth),
		}
	}

	if options.BackgroundColor == nil {
		options.BackgroundColor = theme.DisabledButtonColor()
	}

	if options.LineColor == nil {
		options.LineColor = lineColor
	}

	if options.Scheme == nil {
		options.Scheme = AnalogousScheme(nil)
	}

	plot.options = options
	plot.BaseWidget.ExtendBaseWidget(plot)
	return plot
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
//
// Implements: fyne.Widget
func (plot *Plot) CreateRenderer() fyne.WidgetRenderer {
	switch plot.kind {
	case PlotBar:
		return newBarChartRenderer(plot)
	case PlotLine:
		return newLineChartRenderer(plot)
	case PlotPie:
		return newPieChartRenderer(plot)
	}
	return newBarChartRenderer(plot)
}

// GetYAt returns the value at the given x position.
func (plot *Plot) GetDataAt(fyne.PointEvent) float32 {
	// TODO
	return 0
}

// SetData set all the data for the chart. Because line and bar charts can stack several data lines, the data is a
// 2 dimensional slice.
func (plot *Plot) SetData(data [][]float32) {
	plot.locker.Lock()
	defer plot.locker.Unlock()
	plot.data = data
	plot.Refresh()
}

// Plot adds a new data line to the chart and draw it.
// Warning, PieChart will ignore the added data, this is only used by
// LineChart and BarChart.
func (plot *Plot) Plot(data []float32) {
	plot.locker.Lock()
	defer plot.locker.Unlock()
	if data != nil && len(data) >= 0 {
		plot.data = append(plot.data, data)
	}
	plot.Refresh()
}

// Clear removes all the data from the chart. It will not Refresh the chart view.
func (plot *Plot) Clear() {
	plot.locker.Lock()
	defer plot.locker.Unlock()
	plot.data = [][]float32{}
}
