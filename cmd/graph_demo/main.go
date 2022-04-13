package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/x/fyne/widget/charts"
)

func main() {
	n := 3
	app := app.NewWithID("test.fynex.metal3d.org")
	w := app.NewWindow("Graphs")

	// to pause the "animation"
	pause := make([]bool, n)

	graphWidgets := make([]*charts.LineChart, n)
	graphBoxes := make([]fyne.CanvasObject, n)
	datas := make([][]float64, n)

	// create n graphs
	for i := range graphWidgets {
		graphWidgets[i] = charts.NewLineChart(nil)
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

		// we will play with mouse event and use drawing zone
		func(j int, g *charts.LineChart) {
			g.OnMouseIn = func(e *desktop.MouseEvent) {
				pause[j] = true
			}

			g.OnMouseMoved = func(e *desktop.MouseEvent) {
				onMouseMoved(g, e)
			}

			g.OnMouseOut = func() {
				pause[j] = false

				// cleanup the drawing zone
				drawZone := g.GetDrawable()
				drawZone.Objects = []fyne.CanvasObject{}
				g.Refresh()
			}
		}(j, g)
	}

	// Let's change the graph colors
	graphWidgets[1].GetOptions().StrokeColor = theme.PrimaryColor()
	graphWidgets[2].GetOptions().StrokeColor = theme.FocusColor()
	//graphWidgets[2].GetOptions().FillColor = theme.PressedColor()

	go func() {
		// Contiuously update the data

		// remove the first data point and add a new one each 500ms
		lock := make([]sync.Mutex, n)
		for range time.Tick(500 * time.Millisecond) {
			for i := range datas {
				if pause[i] {
					continue
				}
				lock[i].Lock()
				datas[i] = append(datas[i][1:], rand.Float64()*50)
				graphWidgets[i].SetData(datas[i])
				lock[i].Unlock()
			}
		}
	}()

	// make a sinusoidal graph
	sinus := charts.NewLineChart(nil)

	// set the number of value to plot
	const nx = 100
	// set the y values slice
	siny := make([]float64, nx)

	for i := range [nx]int{} {
		siny[i] = math.Sin(float64(i) / 10) // devide per 100 to get a smooth curve
	}
	sinus.SetData(siny)

	// create a slider values
	slider := widget.NewSlider(float64(nx), 1000)
	slider.OnChanged = func(value float64) {
		siny = make([]float64, int(value))
		for i := 0; i < int(value); i++ {
			siny[i] = math.Sin(float64(i) / 10) // devide per 100 to get a smooth curve
		}
		sinus.SetData(siny)
	}
	sinContainer := container.NewBorder(
		nil, slider, nil, nil, sinus,
	)

	// build the UI
	grid := container.NewGridWithColumns(2, graphBoxes...)
	grid.Add(sinContainer)

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

// This function is called when the mouse is moved on the graph. It creates 2 lines + on circle and a text to display the value at the mouse position.
func onMouseMoved(g *charts.LineChart, e *desktop.MouseEvent) {
	// get the value of data at the mouse position
	val, curvePos := g.GetDataPosAt(e.Position)

	// prepare the vertical verticalLine to display
	lineColor := theme.DisabledColor()
	verticalLine := canvas.NewLine(lineColor)
	verticalLine.Position1 = fyne.NewPos(curvePos.X, 0)
	verticalLine.Position2 = fyne.NewPos(curvePos.X, g.Size().Height)

	horizontalLine := canvas.NewLine(lineColor)
	horizontalLine.Position1 = fyne.NewPos(0, curvePos.Y)
	horizontalLine.Position2 = fyne.NewPos(g.Size().Width, curvePos.Y)

	// place a circle on the curvePos
	circle := canvas.NewCircle(theme.ForegroundColor())
	circle.Resize(fyne.NewSize(theme.TextSize()*.5, theme.TextSize()*.5))
	circle.Move(curvePos.Subtract(fyne.NewPos(
		circle.Size().Height/2, circle.Size().Width/2,
	)))

	// display the value over the circle
	text := canvas.NewText(fmt.Sprintf("%.02f", val), theme.ForegroundColor())
	text.TextSize = theme.TextSize() * 0.7
	text.Move(curvePos.Add(fyne.NewPos(
		text.TextSize*1.5, -text.TextSize*0.5,
	)))

	// then add line, circle and text to the graph
	drawZone := g.GetDrawable()
	drawZone.Objects = []fyne.CanvasObject{verticalLine, horizontalLine, circle, text}
	g.Refresh()

}
