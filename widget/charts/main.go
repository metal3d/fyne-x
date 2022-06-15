// charts package propose some chart widgets.
package charts // import "fyne.io/x/fyne/widget/charts"

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/srwiley/rasterx"
)

// Rasterizer represents a chart that uses a canvas.Raster to draw the chart.
type Rasterizer interface {
	Image() *canvas.Raster
}

// Pointable is a chart where a "pointer event" can be used
// to get the data at a given position.
type Pointable interface {
	// AtPointer return the entire data set and position in
	// the chart at the given pointer position.
	AtPointer(fyne.PointEvent) []Point
}

type Overlayable interface {
	Overlay() *fyne.Container
}

// Point represent a point in the drawn chart.
type Point struct {
	// Value is the "pointed" value
	Value float32
	// Position is the position in the chart where the value is drawn.
	Position fyne.Position
}

// createScanner return a rasterx.ScannerGV for a given width and height.
func createScanner(w, h int) *rasterx.ScannerGV {
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, im, im.Bounds())
	return scanner
}

// createStroker returns a rasterx.Stroker for a given scanner.
func createStroker(scanner *rasterx.ScannerGV) *rasterx.Stroker {
	w, h := scanner.Dest.Bounds().Dx(), scanner.Dest.Bounds().Dy()
	stroker := rasterx.NewStroker(w, h, scanner)
	return stroker
}

// createFiller creates a rasterx.Filler for a given scanner.
func createFiller(scanner *rasterx.ScannerGV) *rasterx.Filler {
	w, h := scanner.Dest.Bounds().Dx(), scanner.Dest.Bounds().Dy()
	filler := rasterx.NewFiller(w, h, scanner)
	return filler
}

// getMin returns the minimum value of a data line.
func getMin(data []float32) float32 {
	min := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
	}
	return min
}

// getMax returns the maximum value of a data line.
func getMax(data []float32) float32 {
	max := data[0]
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	return max
}

// return the reduction factor to apply to data to draw the bar inside the image
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

// return the scale factor applied to Fyne window.
func scaling(o fyne.CanvasObject) float32 {
	if fyne.CurrentApp().Driver() == nil ||
		fyne.CurrentApp().Driver().CanvasForObject(o) == nil {
		return 1
	}
	return fyne.CurrentApp().Driver().CanvasForObject(o).Scale()
}

// find the best zeroY for a complete data set
func globalZeroAxisY(plot *Chart, im image.Image) (zeroY, scaler float64) {
	for _, d := range plot.data {
		scale := scaleData(d, im)
		if scaler == 0 || scaler > scale {
			scaler = scale
			zeroY = zeroAxisY(d, im)
		}
	}
	return
}

// largerDataLine return the longest data line size.
func largerDataLine(data [][]float32) float64 {
	var size int
	for _, d := range data {
		if len(d) > size {
			size = len(d)
		}
	}
	return float64(size)
}
