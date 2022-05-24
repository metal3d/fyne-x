package main

import (
	"image/color"
	"math/rand"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

func main() {
	app := app.New()
	win := app.NewWindow("Framed Demo")
	win.Resize(fyne.NewSize(800, 600))

	// Simple framed, no border radius
	content := xwidget.NewFramed(createContent(), nil)

	// a more complex frame, here we set a border radius, border color
	// and gradient background...
	form := xwidget.NewFramed(createForm(), &xwidget.FramedOptions{
		BorderRadius: theme.TextSize(),
		StrokeWidth:  2,
		StrokeColor:  theme.ForegroundColor(),
		BackgroundGradient: &xwidget.FramedGradient{
			ColorSteps: xwidget.FramedGradientStep{
				0: color.Transparent,
				1: theme.PrimaryColor(),
			},
		},
	})

	// top box with the label and image + the form
	topBox := container.NewGridWithColumns(2, content, form)

	// a grid to add random framed blocks
	frames := container.NewGridWithColumns(2)
	scrollBox := container.NewVScroll(frames)

	// Buttons to create some frames in the grid
	count := 0
	button := widget.NewButton("Add simple Frame", func() {
		count++
		frame := createFrame(count, false)
		frames.Add(frame)
		frames.Refresh()
		scrollBox.ScrollToBottom()
	})
	graduentButton := widget.NewButton("Add gradient Frame", func() {
		count++
		frame := createFrame(count, true)
		frames.Add(frame)
		frames.Refresh()
		scrollBox.ScrollToBottom()
	})

	// initialize some random frames
	for i := 0; i < 5; i++ {
		frame := createFrame(i, i%2 != 0)
		frames.Add(frame)
		count = i
	}

	win.SetContent(container.NewBorder(
		topBox,
		container.NewHBox(button, graduentButton),
		nil,
		nil,
		container.NewBorder(nil, nil, nil, nil, scrollBox),
	))

	win.ShowAndRun()
}

// This function creates some framed label with gradient or simple color background.
func createFrame(i int, gradient bool) fyne.CanvasObject {

	// create a random color
	r := uint8(rand.Intn(255))
	g := uint8(rand.Intn(255))
	b := uint8(rand.Intn(255))
	c := color.RGBA{r, g, b, 255}

	opts := &xwidget.FramedOptions{
		BackgroundColor: c,
		BorderRadius:    24,
	}
	if gradient {
		// create a gradient from top Transparent to bottom background color
		opts.BackgroundGradient = &xwidget.FramedGradient{
			ColorSteps: xwidget.FramedGradientStep{
				0: color.Transparent,
				1: c,
			},
			Direction: xwidget.GradientDirectionTopDown,
		}
	}

	// show an hello world + the background color
	label := widget.NewLabel("Hello World " + strconv.Itoa(i+1) + "\n" +
		"R: " + strconv.Itoa(int(r)) +
		" G: " + strconv.Itoa(int(g)) +
		" B: " + strconv.Itoa(int(b)))

	frame := xwidget.NewFramed(label, opts)
	return frame
}

func createContent() fyne.CanvasObject {
	label := widget.NewLabel("This block is in a frame")
	logo := theme.FyneLogo()
	im := canvas.NewImageFromResource(logo)
	im.FillMode = canvas.ImageFillOriginal
	im.SetMinSize(fyne.NewSize(80, 80))

	return container.NewVBox(
		label,
		im,
	)
}

func createForm() fyne.CanvasObject {
	title := widget.NewLabel("Form example")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	items := []*widget.FormItem{
		{Text: "Name", Widget: widget.NewEntry()},
		{Text: "Age", Widget: widget.NewEntry()},
		{Text: "Email", Widget: widget.NewEntry()},
	}
	form := widget.NewForm(items...)
	return container.NewVBox(title, form)
}
