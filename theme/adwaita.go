package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// AdwaitaDark scheme
var AdwaitaDarkScheme = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.RGBA{0x24, 0x24, 0x24, 0xff},
	theme.ColorNameForeground:      color.RGBA{0xff, 0xff, 0xff, 0xff},
	theme.ColorNamePrimary:         color.RGBA{0x78, 0xae, 0xed, 0xff},
	theme.ColorNameInputBackground: color.RGBA{0x30, 0x30, 0x30, 0xff},
	theme.ColorNameDisabled:        color.RGBA{0x30, 0x30, 0x30, 0xff},
	theme.ColorNameError:           color.RGBA{0xc0, 0x1c, 0x28, 0xff},
}

// AdwaitaLight scheme
var AdwaitaLightScheme = map[fyne.ThemeColorName]color.Color{
	theme.ColorNameBackground:      color.RGBA{0xfa, 0xfa, 0xfa, 0xfa},
	theme.ColorNameForeground:      color.RGBA{0x0, 0x0, 0x0, 0xff},
	theme.ColorNamePrimary:         color.RGBA{0x1c, 0x71, 0xd8, 0xff},
	theme.ColorNameInputBackground: color.RGBA{0xfa, 0xfa, 0xfa, 0xff},
	theme.ColorNameDisabled:        color.RGBA{0xfa, 0xfa, 0xfa, 0xff},
	theme.ColorNameError:           color.RGBA{0xe0, 0x1b, 0x24, 0xff},
}

var _ fyne.Theme = (*Adwaita)(nil)

// Adwaita is a theme that follows the Adwaita theme.
// See: https://gnome.pages.gitlab.gnome.org/libadwaita/doc/main/named-colors.html
type Adwaita struct {
	override fyne.Theme
}

// NewAdwaita returns a new Adwaita theme.
func NewAdwaita() fyne.Theme {
	return &Adwaita{}
}

// Color returns the named color for the current theme.
func (a *Adwaita) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch variant {
	case theme.VariantLight:
		if c, ok := AdwaitaLightScheme[name]; ok {
			return c
		}
	case theme.VariantDark:
		if c, ok := AdwaitaDarkScheme[name]; ok {
			return c
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Font returns the named font for the current theme.
func (a *Adwaita) Font(style fyne.TextStyle) fyne.Resource {
	if a.override != nil {
		return a.override.Font(style)
	}
	return theme.DefaultTheme().Font(style)
}

// Icon returns the named resource for the current theme.
func (a *Adwaita) Icon(name fyne.ThemeIconName) fyne.Resource {
	if a.override != nil {
		return a.override.Icon(name)
	}
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size of the named resource for the current theme.
func (a *Adwaita) Size(name fyne.ThemeSizeName) float32 {
	if a.override != nil {
		return a.override.Size(name)
	}
	return theme.DefaultTheme().Size(name)
}

// setGTKFallbackTheme provides a way to override the theme with a custom one. Actually, it's a hack to
// get Gnome specific sizes, icons and fonts using FromDesktopEnvironment() function.
func (a *Adwaita) setGTKFallbackTheme(t fyne.Theme) {
	a.override = t
}
