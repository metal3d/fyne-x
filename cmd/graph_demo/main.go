package main

import (
	"fmt"
	"log"
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
	datas := make([][]float32, n)

	// create n graphs
	for i := range graphs {
		graphs[i] = xwidget.NewGraph(xwidget.GraphOpts{
			Title: xwidget.GraphTitle{
				Text: fmt.Sprintf("Graph %d", i+1),
			},
		})
	}
	for j := range datas {
		// fill the data with random values
		datas[j] = make([]float32, 64)
		for k := range datas[j] {
			datas[j][k] = rand.Float32()*50 - 25
		}
		graphs[j].SetData(datas[j])

		// we will play with mouse event and use drawing zone
		func(j int) {
			graphs[j].OnMouseIn = func(e *desktop.MouseEvent) {
				log.Println("pause on graph", j+1)
				pause[j] = true
			}

			graphs[j].OnMouseMoved = func(e *desktop.MouseEvent) {

				// get the value of data at the mouse position
				val, curvePos := graphs[j].GetDataPosAt(e.Position)

				// prepare the vertical line to display
				line := canvas.NewLine(theme.ForegroundColor())
				line.Position1 = fyne.NewPos(curvePos.X, 0)
				line.Position2 = fyne.NewPos(curvePos.X, graphs[j].Size().Height)

				// place a circle on the curvePos
				circle := canvas.NewCircle(theme.ForegroundColor())
				circle.Resize(fyne.NewSize(theme.TextSize(), theme.TextSize()))
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
				drawZone.Objects = []fyne.CanvasObject{line, circle, text}
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

	// Change some graph title. We can do it with NewGraph(opts) but that's an example here

	go func() {
		lock := make([]sync.Mutex, n)
		// remove the first data point and add a new one each 500ms
		for range time.Tick(500 * time.Millisecond) {
			for i := range datas {
				if pause[i] {
					continue
				}
				lock[i].Lock()
				datas[i] = append(datas[i][1:], rand.Float32()*50-25)
				graphs[i].SetData(datas[i])
				lock[i].Unlock()
			}
		}
	}()

	grid := container.NewGridWithColumns(2)
	for _, graph := range graphs {
		grid.Add(graph)
	}
	content := container.NewBorder(widget.NewLabel("Graphs"), nil, nil, nil,
		grid,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(580, 340))
	w.ShowAndRun()
}
