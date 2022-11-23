//go:build linux
// +build linux

package theme

import (
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/x/fyne/theme/desktop"
)

// FromDesktopEnvironment returns a new WindowManagerTheme instance for the current desktop session.
// If the desktop manager is not supported or if it is not found, return the default theme
// Flags are optional, they define which desktop settings to use. If set to DesktopGrapNone, so only the colors are applied.
func FromDesktopEnvironment(flags ...DesktopGrapFlag) fyne.Theme {
	wm := os.Getenv("XDG_CURRENT_DESKTOP")
	if wm == "" {
		wm = os.Getenv("DESKTOP_SESSION")
	}
	wm = strings.ToLower(wm)

	switch wm {
	case "gnome", "gnome-shell", "unity", "gnome-classic", "ubuntu:gnome", "ubuntu:unity":
		adw := NewAdwaita()
		adw.(*Adwaita).setGTKFallbackTheme(desktop.NewGTKTheme(-1, flags...))
		return adw
	case "xfce", "mate", "gnome-mate":
		return desktop.NewGTKTheme(-1, flags...)
	case "kde", "kde-plasma", "plasma", "lxqt":
		return desktop.NewKDETheme()

	}

	log.Println("Window manager not supported:", wm, "using default theme")
	return theme.DefaultTheme()
}
