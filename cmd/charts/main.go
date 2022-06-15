package main

import (
	"image/color"
	"log"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"fyne.io/x/fyne/widget/charts"
)

var Pause bool

func main() {
	app := app.New()
	win := app.NewWindow("Charts")
	win.Resize(fyne.NewSize(800, 300))

	N := 40
	data := make([]float32, N)

	for i := 0; i < N; i++ {
		data[i] = rand.Float32()*20 - 10
	}

	line := charts.NewChart(charts.Line, &charts.Options{
		LineWidth: 2,
		//BackgroundColor: color.RGBA{0x00, 0x00, 0x33, 0x00},
	})
	line.Plot(data)

	bar := charts.NewChart(charts.Bar, nil)
	bar.Plot(data)

	pie := charts.NewChart(charts.Pie, nil)
	pie.Plot([]float32{30, 20, 55, 34})

	pie2 := charts.NewChart(charts.Pie, &charts.Options{
		Scheme:    charts.AnalogousScheme(color.RGBA{0xAE, 0x44, 0x44, 0x00}),
		LineWidth: 2,
	})
	pie2.Plot([]float32{45, 5, 33, 17})

	N = 16
	multidata := make([][]float32, 3)
	for i := range multidata {
		multidata[i] = make([]float32, N)
		for j := 0; j < N; j++ {
			multidata[i][j] = (rand.Float32()*20 - 10) / float32(i+1)
		}
	}

	multiline := NewMousePlot(charts.Line, nil)
	multiline2 := charts.NewChart(charts.Line, &charts.Options{
		LineWidth: 1,
		Scheme:    charts.RandomScheme(),
	})
	multiline.SetData(multidata)
	multiline2.SetData(multidata)

	// a multi bar data
	multibar := charts.NewChart(charts.Bar, nil)
	multibar.Plot([]float32{4, -6, 7, 8, 2, -1, 6})
	multibar.Plot([]float32{6, -3, 2, 3, 4, -2, 3})

	grid := container.NewGridWithColumns(3,
		line, bar, pie, pie2, multiline, multiline2, multibar,
	)

	go func() {
		for {
			time.Sleep(time.Millisecond * 400)
			if Pause {
				continue
			}
			d := rand.Float32()*20 - 10
			data = append(data[1:], d)
			bar.SetData([][]float32{data})
			line.SetData([][]float32{data})
			bar.Refresh()
			line.Refresh()

			for i := range multidata {
				d = (rand.Float32()*20 - 10) / float32(i+1)
				multidata[i] = append(multidata[i][1:], d)
			}
			multiline.SetData(multidata)
			multiline2.SetData(multidata)

		}
	}()
	win.SetContent(grid)
	win.ShowAndRun()
}

var _ desktop.Hoverable = (*MousePlot)(nil)
var _ fyne.Widget = (*MousePlot)(nil)

type MousePlot struct {
	widget.BaseWidget
	*charts.Chart
}

func NewMousePlot(kind charts.Type, opts *charts.Options) *MousePlot {
	m := &MousePlot{Chart: charts.NewChart(kind, opts)}
	m.BaseWidget.ExtendBaseWidget(m)
	return m
}

func (m *MousePlot) Refresh() {
	m.Chart.Refresh()
}

func (m *MousePlot) MouseIn(e *desktop.MouseEvent) {
	Pause = true
	log.Println("MouseIn")
}

func (m *MousePlot) MouseMoved(e *desktop.MouseEvent) {
	p := m.GetDataAt(e.PointEvent)
	circles := []fyne.CanvasObject{}
	for _, p := range p {
		circle := canvas.NewCircle(color.White)
		circle.FillColor = color.White
		circle.Resize(fyne.NewSize(10, 10))
		circle.Move(p.Position.Add(fyne.NewPos(5, 0)))
		circles = append(circles, circle)
	}

	overlay := m.Overlay()
	if overlay == nil {
		return
	}
	overlay.Objects = circles
	m.Refresh()

}

func (m *MousePlot) MouseOut() {
	Pause = false
	overlay := m.Overlay()
	if overlay == nil {
		return
	}
	overlay.Objects = []fyne.CanvasObject{}
	m.Refresh()
}
