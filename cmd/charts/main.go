package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
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

	line := charts.NewLineChart(&charts.Options{
		LineWidth:       2,
		BackgroundColor: color.RGBA{0x00, 0x00, 0x33, 0x00},
	})
	line.Plot(data)

	bar := charts.NewBarChart(nil)
	bar.Plot(data)

	pie := charts.NewPieChart(nil)
	pie.Plot([]float32{30, 20, 55, 34})

	//pie2 := charts.NewChart(charts.Pie, &charts.Options{
	pie2 := NewMousePie(&charts.Options{
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
	multiline2 := charts.NewLineChart(&charts.Options{
		LineWidth: 1,
		Scheme:    charts.RandomScheme(),
	})
	multiline.SetData(multidata)
	multiline2.SetData(multidata)

	// a multi bar data
	multibar := NewMousePlot(charts.Bar, nil)
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
	charts.Chart
	kind charts.Type
}

func NewMousePlot(kind charts.Type, opts *charts.Options) *MousePlot {
	m := &MousePlot{Chart: *charts.NewChart(kind, opts)}
	m.kind = kind
	m.Debug = true
	return m
}

//func (m *MousePlot) Refresh() {
//	m.Chart.Refresh()
//}

func (m *MousePlot) MouseIn(e *desktop.MouseEvent) {
	Pause = true
}

func (m *MousePlot) MouseMoved(e *desktop.MouseEvent) {
	overlay := m.Overlay()
	if overlay == nil {
		return
	}
	points := m.GetDataAt(e.PointEvent)
	if points == nil {
		return
	}
	draw := []fyne.CanvasObject{}
	for i, p := range points {

		text := canvas.NewText(fmt.Sprintf("%0.2f", p.Value), m.Options().Scheme[i])
		text.TextStyle.Bold = true

		var tx, ty float32
		if m.kind == charts.Line {
			// for line chart, draw lines that hits the point
			// at the given pointer poisition

			//make a nice color
			r, g, b, a := m.Options().Scheme[i].RGBA()
			a = 0x80

			line := canvas.NewLine(color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
			line.Position1 = fyne.NewPos(0, p.Position.Y)
			line.Position2 = fyne.NewPos(m.Size().Width, p.Position.Y)
			draw = append(draw, line)

			// make the text to not overflow the image
			text.TextSize = 9
			s := fyne.MeasureText(text.Text, text.TextSize, text.TextStyle)
			ty = -s.Height / 2
			tx = 5
			if p.Position.X > m.Size().Width/2 {
				tx = -s.Width - tx
			}
		} else {
			// for bar chart, draw the text inside the bar
			text.TextSize = 7
			s := fyne.MeasureText(text.Text, text.TextSize, text.TextStyle)
			tx = -s.Width / 2
			if p.Value < 0 {
				ty = -s.Height
			}
		}
		text.Move(p.Position.Add(fyne.NewPos(tx, ty)))
		draw = append(draw, text)
	}

	if m.kind == charts.Line {
		// a vertical line to mark the mouse position on line charts. We only use the
		// first point to know the position.
		col := color.NRGBA{0xff, 0xff, 0xff, 0x88}
		vline := canvas.NewLine(col)
		vline.Position1 = fyne.NewPos(points[0].Position.X, 0)
		vline.Position2 = fyne.NewPos(points[0].Position.X, m.Size().Height)
		draw = append(draw, vline)
	}

	overlay.Objects = draw
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

var _ desktop.Hoverable = (*MousePie)(nil)
var _ fyne.Widget = (*MousePie)(nil)

type MousePie struct {
	*charts.Chart
}

func NewMousePie(opts *charts.Options) *MousePie {
	m := &MousePie{Chart: charts.NewPieChart(opts)}
	m.BaseWidget.ExtendBaseWidget(m)
	return m
}

func (m *MousePie) Refresh() {
	m.Chart.Refresh()
}

func (m *MousePie) MouseIn(e *desktop.MouseEvent) {}

func (m *MousePie) MouseMoved(e *desktop.MouseEvent) {
	overlay := m.Overlay()
	if overlay == nil {
		return
	}
	points := m.GetDataAt(e.PointEvent)
	if len(points) == 0 {
		overlay.Objects = []fyne.CanvasObject{}
		return
	}

	p := points[0]
	text := canvas.NewText(fmt.Sprintf("%0.1f%%", p.Value), color.White)
	text.TextSize = 7
	s := fyne.MeasureText(text.Text, text.TextSize, text.TextStyle)
	text.Move(p.Position.Subtract(fyne.NewPos(s.Width/2, s.Height/2)))

	overlay.Objects = []fyne.CanvasObject{text}
	m.Refresh()
}

func (m *MousePie) MouseOut() {
	overlay := m.Overlay()
	if overlay == nil {
		return
	}
	overlay.Objects = []fyne.CanvasObject{}
}
