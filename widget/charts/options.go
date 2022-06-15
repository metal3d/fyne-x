package charts

import "image/color"

// Options holds the options for the plot.
type Options struct {
	// LineWidth is the width of the line. Default to 1.0.
	LineWidth float32
	// BackgroundColor is the background that fills the "line" chart (below the line).
	BackgroundColor color.Color
	// LineColor is the color of the line. Default is color.Transparent.
	LineColor color.Color
	// Scheme is the color scheme for the chart.
	// Note: Pie chart uses this to define the colors of the data.
	// Default is a generated color scheme from the primary color of the theme.
	Scheme Scheme
}
