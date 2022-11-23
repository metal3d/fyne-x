package desktop

// DesktopGrapFlag is a flag to indicate some desktop settings to use
type DesktopGrapFlag uint8

// DesktopGrapNone is a flag to indicate that the theme should not use the desktop settings
const DesktopGrapNone DesktopGrapFlag = 0

const (
	// DesktopGrabFonts is a flag to indicate that the theme should use the desktop font settings
	DesktopGrabFonts DesktopGrapFlag = 1 << iota
	// DesktopGrabIcons is a flag to indicate that the theme should use the desktop icons theme
	DesktopGrabIcons
	// DesktopGrapScale is a flag to indicate that the theme should use the desktop size settings
	DesktopGrapScale
	// DesktopGrabAll is a flag to indicate that the theme should use all desktop settings
	DesktopGrabAll = DesktopGrabFonts | DesktopGrabIcons | DesktopGrapScale
)
