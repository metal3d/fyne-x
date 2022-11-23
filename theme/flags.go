package theme

import "fyne.io/x/fyne/theme/desktop"

// DesktopGrapFlag is a flag to indicate some desktop settings to use - alisas for desktop.DesktopGrapFlag.
type DesktopGrapFlag = desktop.DesktopGrapFlag

const (
	// DesktopGrabNone is a flag to indicate that the theme should not use the desktop settings - alisas for desktop.DesktopGrabNone.
	DesktopGrabNone DesktopGrapFlag = desktop.DesktopGrapNone
	// DesktopGrabFonts is a flag to indicate that the theme should use the desktop font settings
	DesktopGrabFonts DesktopGrapFlag = desktop.DesktopGrabFonts
	// DesktopGrabIcons is a flag to indicate that the theme should use the desktop icons theme
	DesktopGrabIcons DesktopGrapFlag = desktop.DesktopGrabIcons
	// DesktopGrabScale is a flag to indicate that the theme should use the desktop size settings
	DesktopGrabScale DesktopGrapFlag = desktop.DesktopGrapScale
	// DesktopGrabAll is a flag to indicate that the theme should use all desktop settings
	DesktopGrabAll DesktopGrapFlag = desktop.DesktopGrabAll
)
