package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/x/fyne/widget/charts"
)

func main() {
	app := app.New()
	w := app.NewWindow("Graphs")

	// to pause the "animation"

	numberOfLineChart := 2
	graphWidgets := make([]*LineChartWithMouse, numberOfLineChart)
	graphBoxes := make([]fyne.CanvasObject, numberOfLineChart)
	datas := make([][]float64, numberOfLineChart)

	// create n graphs
	for i := range graphWidgets {
		graphWidgets[i] = NewLineChartWithMouse()
		// Set a title for the graph, use nice Border layout
		graphBoxes[i] = container.NewBorder(
			widget.NewLabelWithStyle(fmt.Sprintf("Graph %d", i+1), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			graphWidgets[i])
	}

	for j, g := range graphWidgets {
		// fill the data with random values
		datas[j] = make([]float64, 64)
		for k := range datas[j] {
			datas[j][k] = rand.Float64() * 50
		}
		g.SetData(datas[j])

	}

	// Let's change the graph colors
	graphWidgets[1].GetOptions().StrokeColor = theme.PrimaryColor()

	go func() {
		// Contiuously update the data

		// remove the first data point and add a new one each 500ms
		lock := make([]sync.Mutex, numberOfLineChart)
		for range time.Tick(500 * time.Millisecond) {
			for i := range datas {
				if graphWidgets[i].IsMouseOver() {
					continue
				}
				lock[i].Lock()
				datas[i] = append(datas[i][1:], rand.Float64()*50)
				graphWidgets[i].SetData(datas[i])
				lock[i].Unlock()
			}
		}
	}()
	grid := container.NewGridWithColumns(2, graphBoxes...)

	sinContainer := NewSliderGraph(100, 1000, 10)
	grid.Add(sinContainer.Container())

	// create a barchart
	bar := charts.NewHistrogramChart(nil)
	bardata := make([]float64, 10)
	for i := range bardata {
		bardata[i] = rand.Float64() * 10
	}
	bar.GetOptions().FillColor = theme.FocusColor()
	bar.SetData(bardata)
	grid.Add(bar)

	go func() {
		for range time.Tick(700 * time.Millisecond) {
			bardata = append(bardata[1:], rand.Float64()*10)
			bar.SetData(bardata)
		}
	}()

	md := widget.NewLabel(`
>> Graphs Demo

This is a simple example of what you can do with graphs.
Pass mouse on the dynamic graphs to see some
"OnMouseMoved" behaviors.

On the left, it's a simple sinusoidal function drawn.
    `)
	grid.Add(md)

	w.SetContent(grid)
	w.Resize(fyne.NewSize(580, 340))
	w.ShowAndRun()
}
