package charts

// NewPieChart creates a new pie chart. It's an alias on NewChart(Pie, options).
func NewPieChart(options *Options) *Chart {
	return NewChart(Pie, options)
}

// NewLineChart creates a new line chart. It's an alias on NewChart(Line, options).
func NewLineChart(options *Options) *Chart {
	return NewChart(Line, options)
}

// NewBarChart creates a new bar chart. It's an alias on NewChart(Bar, options).
func NewBarChart(options *Options) *Chart {
	return NewChart(Bar, options)
}
