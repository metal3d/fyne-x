package main

import (
	"image/color"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/x/fyne/widget/charts"
)

func main() {
	app := app.New()
	win := app.NewWindow("Charts")
	win.Resize(fyne.NewSize(800, 300))

	N := 40
	data := make([]float32, N)

	for i := 0; i < N; i++ {
		data[i] = rand.Float32()*20 - 10
	}

	line := charts.NewPlot(charts.PlotLine, &charts.PlotOptions{
		LineWidth:       2,
		BackgroundColor: color.RGBA{0x00, 0x00, 0x33, 0x00},
	})
	line.Plot(data)

	bar := charts.NewPlot(charts.PlotBar, nil)
	bar.Plot(data)

	pie := charts.NewPlot(charts.PlotPie, nil)
	pie.Plot([]float32{30, 20, 55, 34})

	pie2 := charts.NewPlot(charts.PlotPie, &charts.PlotOptions{
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

	multiline := charts.NewPlot(charts.PlotLine, nil)
	multiline2 := charts.NewPlot(charts.PlotLine, &charts.PlotOptions{
		Scheme: charts.RandomScheme(),
	})
	multiline.SetData(multidata)
	multiline2.SetData(multidata)

	// a multi bar data
	multibar := charts.NewPlot(charts.PlotBar, nil)
	multibar.Plot([]float32{4, -6, 7, 8, 2, -1, 6})
	multibar.Plot([]float32{6, -3, -8, 3, 4, -2, 3})

	grid := container.NewGridWithColumns(3,
		line, bar, pie, pie2, multiline, multiline2, multibar,
	)

	go func() {
		for {
			time.Sleep(time.Millisecond * 400)
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
