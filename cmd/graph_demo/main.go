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
	xwidget "fyne.io/x/fyne/widget"
)

func main() {
	n := 4
	app := app.New()
	w := app.NewWindow("Graphs")

	// to pause the "animation"
	pause := make([]bool, n)

	graphs := make([]*xwidget.Graph, n)
	graphbox := make([]fyne.CanvasObject, n)
	datas := make([][]float32, n)

	// create n graphs
	for i := range graphs {
		graphs[i] = xwidget.NewGraph(nil)
		graphbox[i] = container.NewBorder(
			widget.NewLabelWithStyle(fmt.Sprintf("Graph %d", i+1), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			graphs[i])
	}
	for j := range datas {
		// fill the data with random values
		datas[j] = make([]float32, 64)
		for k := range datas[j] {
			datas[j][k] = rand.Float32() * 50
		}
		graphs[j].SetData(datas[j])

		// we will play with mouse event and use drawing zone
		func(j int) {
			graphs[j].OnMouseIn = func(e *desktop.MouseEvent) {
				pause[j] = true
			}

			graphs[j].OnMouseMoved = func(e *desktop.MouseEvent) {

				// get the value of data at the mouse position
				val, curvePos := graphs[j].GetDataPosAt(e.Position)

				// prepare the vertical verticalLine to display
				lineColor := theme.DisabledColor()
				verticalLine := canvas.NewLine(lineColor)
				verticalLine.Position1 = fyne.NewPos(curvePos.X, 0)
				verticalLine.Position2 = fyne.NewPos(curvePos.X, graphs[j].Size().Height)

				horizontalLine := canvas.NewLine(lineColor)
				horizontalLine.Position1 = fyne.NewPos(0, curvePos.Y)
				horizontalLine.Position2 = fyne.NewPos(graphs[j].Size().Width, curvePos.Y)

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
				drawZone := graphs[j].GetDrawable()
				drawZone.Objects = []fyne.CanvasObject{verticalLine, horizontalLine, circle, text}
				graphs[j].Refresh()
			}

			graphs[j].OnMouseOut = func() {
				pause[j] = false

				// cleanup the drawing zone
				drawZone := graphs[j].GetDrawable()
				drawZone.Objects = []fyne.CanvasObject{}
				graphs[j].Refresh()
			}
		}(j)
	}

	// Let's change the graph colors
	graphs[1].GetOptions().StrokeColor = theme.PrimaryColor()
	graphs[2].GetOptions().StrokeColor = theme.FocusColor()
	graphs[2].GetOptions().FillColor = theme.PressedColor()

	go func() {
		lock := make([]sync.Mutex, n)
		// remove the first data point and add a new one each 500ms
		for range time.Tick(500 * time.Millisecond) {
			for i := range datas {
				if pause[i] {
					continue
				}
				lock[i].Lock()
				datas[i] = append(datas[i][1:], rand.Float32()*50)
				graphs[i].SetData(datas[i])
				lock[i].Unlock()
			}
		}
	}()

	// make a sinusoidal graph
	sinus := xwidget.NewGraph(nil)

	const nx = 1024
	siny := make([]float32, nx)

	for i := range [nx]int{} {
		siny[i] = float32(math.Sin(float64(i) / 100))
	}
	sinus.SetData(siny)

	// build the UI
	grid := container.NewGridWithColumns(2, graphbox...)
	grid.Add(sinus)

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
