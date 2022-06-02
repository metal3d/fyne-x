// charts package propose some chart widgets.
package charts // import "fyne.io/x/fyne/widget/charts"

import (
	"image"

	"fyne.io/fyne/v2"
	"github.com/srwiley/rasterx"
)

type Plotable interface {
	// GetDataAt returns the data at the given pointer position.
	GetDataAt(fyne.PointEvent) float32

	// AddData adds a new data set to the chart and plot the chart.
	Plot(data []float32)

	// SetData sets the data to be drawn.
	SetData([][]float32)

	// Clear clears the data.
	Clear()
}

func createScanner(w, h int) *rasterx.ScannerGV {
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, im, im.Bounds())
	return scanner
}

func createStroker(scanner *rasterx.ScannerGV) *rasterx.Stroker {
	w, h := scanner.Dest.Bounds().Dx(), scanner.Dest.Bounds().Dy()
	stroker := rasterx.NewStroker(w, h, scanner)
	return stroker
}

func createFiller(scanner *rasterx.ScannerGV) *rasterx.Filler {
	w, h := scanner.Dest.Bounds().Dx(), scanner.Dest.Bounds().Dy()
	filler := rasterx.NewFiller(w, h, scanner)
	return filler
}

func getMin(data []float32) float32 {
	min := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
	}
	return min
}

func getMax(data []float32) float32 {
	max := data[0]
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	return max
}

// return the ratio of the data to draw the bar inside the image
func scaleData(data []float32, im image.Image) float64 {
	min := getMin(data)
	max := getMax(data)
	h := max - min
	return float64(im.Bounds().Dy()) / float64(h)
}

// find where to place the X axis on Y of the image. Note that the origin is
// at the top left corner of the image. We need to start from the bottom left.
func zeroAxisY(data []float32, im image.Image) float64 {
	min := float64(getMin(data))
	return float64(im.Bounds().Dy()) - (-min * scaleData(data, im))
}

func scaling(o fyne.CanvasObject) float32 {
	if fyne.CurrentApp().Driver() == nil ||
		fyne.CurrentApp().Driver().CanvasForObject(o) == nil {
		return 1
	}
	return fyne.CurrentApp().Driver().CanvasForObject(o).Scale()
}

// find the best zeroY for a complete data set
func globalZeroAxisY(plot *Plot, im image.Image) (zeroY, scaler float64) {
	for _, d := range plot.data {
		scale := scaleData(d, im)
		if scaler == 0 || scaler > scale {
			scaler = scale
			zeroY = zeroAxisY(d, im)
		}
	}
	return
}
