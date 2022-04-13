package charts

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type graph struct {
	// OnMouseIn is trigger when the mouse enters the widget.
	OnMouseIn func(*desktop.MouseEvent)

	// OnMouseOut is trigger when the mouse exits the widget.
	OnMouseOut func()

	// OnMouseMoved is trigger when the mouse moves over the widget.
	OnMouseMoved func(*desktop.MouseEvent)

	// OnMouseUp is trigger when the mouse button is clicked or tapped on mobile device.
	OnTapped func(*fyne.PointEvent)
}

// MouseIn is called when the mouse is moved over the widget.
//
// implements desktop.Hoverable
func (g *graph) MouseIn(e *desktop.MouseEvent) {
	if g.OnMouseIn != nil {
		g.OnMouseIn(e)
	}
}

// MouseMoved is called when the mouse is moved over the widget.
//
// implements desktop.Hoverable
func (g *graph) MouseMoved(e *desktop.MouseEvent) {
	if g.OnMouseMoved != nil {
		g.OnMouseMoved(e)
	}
}

// MouseOut is called when the mouse is moved out of the widget.
//
// implements desktop.Hoverable
func (g *graph) MouseOut() {
	if g.OnMouseOut != nil {
		g.OnMouseOut()
	}
}

// Tapped is called when the widget is tapped or clicked.
//
// implements fyne.Tappable
func (g *graph) Tapped(e *fyne.PointEvent) {
	if g.OnTapped != nil {
		g.OnTapped(e)
	}
}
