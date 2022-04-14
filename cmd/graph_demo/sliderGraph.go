package main

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/x/fyne/widget/charts"
)

type SliderGraph struct {
	*charts.LineChart

	container *fyne.Container
	slider    *widget.Slider

	precision float64
}

func NewSliderGraph(nMin, nMax int, precision float64) *SliderGraph {

	chart := charts.NewLineChart(nil)
	slider := widget.NewSlider(float64(nMin), float64(nMax))
	container := container.NewBorder(nil, slider, nil, nil, chart)

	// use border layout
	sg := &SliderGraph{
		container: container,
		LineChart: chart,
		slider:    slider,
		precision: precision,
	}

	// init all sinus values
	siny := make([]float64, nMin)
	for i := 0; i < nMin; i++ {
		siny[i] = math.Sin(float64(i) / sg.precision)
	}

	sg.LineChart.SetData(siny)
	sg.slider.OnChanged = sg.slided // connect slider event to "slided"

	return sg

}

// Container returns the container of the widget
func (sg *SliderGraph) Container() fyne.CanvasObject {
	return sg.container
}

func (sg *SliderGraph) slided(value float64) {
	siny := make([]float64, int(value))
	for i := 0; i < int(value); i++ {
		siny[i] = math.Sin(float64(i) / float64(sg.precision))
	}
	sg.LineChart.SetData(siny)
}
