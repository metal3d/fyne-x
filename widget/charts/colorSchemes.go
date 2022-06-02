package charts

import (
	"image/color"
	"math/rand"
	"time"

	"fyne.io/fyne/v2/theme"
)

type Scheme []color.Color

var (
	naturalScheme Scheme
	monotone      Scheme
)

func (scheme Scheme) ColorAt(index int) color.Color {
	return scheme[index%len(scheme)]
}

func RandomScheme() Scheme {
	var r Scheme
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 360; i++ {
		col := color.NRGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}
		r = append(r, col)
	}
	return r
}

func AnalogousScheme(base color.Color) Scheme {

	if base == nil {
		base = theme.PrimaryColor()
	}
	scheme := make(Scheme, 0)
	r, g, b, _ := base.RGBA()
	h, c, l := HsluvFromRGB(float64(r&0xff)/255, float64(g&0xff)/255, float64(b&0xff)/255)
	offset := float64(10)
	for i := 0; i < 360; i += 30 {
		// change the luminance
		if i%2 == 0 {
			l = l + offset
		} else {
			l = l - offset
		}
		if l > 100 {
			l = 10
		}
		if l < 0 {
			l = 90
		}
		r, g, b := HsluvToRGB(h, c, l)
		scheme = append(scheme, color.NRGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), 255})
	}
	return scheme
}
