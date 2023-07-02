package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// must be in sync with nord_colors_generator.go - getting the colors from the Nord document page.
//go:generate go run ./nord_colors_generator.go

var _ fyne.Theme = (*Nord)(nil)

// NordTheme returns a new Nord theme.
func NordTheme() fyne.Theme {
	return &Nord{}
}

// Nord is a theme that follows the Nord theme.
// See: https://gnome.pages.gitlab.gnome.org/libnord/doc/main/named-colors.html
type Nord struct{}

// Color returns the named color for the current theme.
func (a *Nord) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch variant {
	case theme.VariantLight:
		if c, ok := nordLightScheme[name]; ok {
			return c
		}
	case theme.VariantDark:
		if c, ok := nordDarkScheme[name]; ok {
			return c
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Font returns the named font for the current theme.
func (a *Nord) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the named resource for the current theme.
func (a *Nord) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size of the named resource for the current theme.
func (a *Nord) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
