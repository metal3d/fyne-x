//go:build !linux
// +build !linux

package theme

import (
	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"
)

// FromDesktopEnvironment returns a new WindowManagerTheme instance for the current desktop session.
// If the desktop manager is not supported or if it is not found, return the default theme
// Flags can be a piped list of DesktopFlags to indicate which desktop settings to use. If set to 0, so only the colors are managed.
func FromDesktopEnvironment(flags ...DesktopFlags) fyne.Theme {
	return fynetheme.DefaultTheme()
}
