package breakapp

import _ "embed"

// menubarIconPNG is the template PNG used as the macOS status-bar icon.
// macOS will auto-tint it for light/dark menu-bar modes (template image).
//
// The 22x22 base size is the standard menu-bar icon target. macOS picks
// the right backing scale based on display DPI; we pass the bytes as-is
// to NSImage which handles the resampling.
//
//go:embed menubar.png
var menubarIconPNG []byte
