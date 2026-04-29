//go:build !darwin

// Package macos provides no-op stubs on non-Darwin platforms so the rest of
// the codebase remains buildable for cross-compilation checks. None of
// these functions do anything useful off macOS.
package macos

import "time"

type MenuTag int
type StatusIconState int

const (
	StatusIconRunning StatusIconState = iota
	StatusIconPaused
	StatusIconBusy
	StatusIconBreak
)

type StatusBar struct{}

func SetupStatusBar(title string) *StatusBar { return &StatusBar{} }

func (s *StatusBar) SetTitle(string)                          {}
func (s *StatusBar) SetBuiltinIcon(StatusIconState)           {}
func (s *StatusBar) SetImage([]byte)                          {}
func (s *StatusBar) AddItem(MenuTag, string, func())          {}
func (s *StatusBar) AddDisabled(MenuTag, string)              {}
func (s *StatusBar) AddSeparator()                            {}
func (s *StatusBar) UpdateItem(MenuTag, string)               {}
func (s *StatusBar) SetItemHidden(MenuTag, bool)              {}
func (s *StatusBar) Teardown()                                {}

type NotifAction int

const (
	NotifDefault NotifAction = iota + 1
	NotifSkip
	NotifPostpone
)

func SetNotificationHandler(func(NotifAction)) {}
func Notify(string, string, bool)              {}

func FrontmostApp() string                  { return "" }
func ActivateApp(string) bool               { return false }
func HideSelfAndActivate(bundleID string)   {}

type IdleSource struct{}

func (IdleSource) IdleFor() time.Duration { return 0 }

type BusyKind uint8

const (
	BusyMicrophone BusyKind = 1 << iota
	BusyMediaPlaying
)

type BusyState uint8

func (s BusyState) IsBusy() bool        { return false }
func (s BusyState) Has(k BusyKind) bool { return false }
func (s BusyState) String() string      { return "" }
func CurrentBusyState() BusyState       { return 0 }

type BusySource struct{}

func (BusySource) BusyState() (string, bool) { return "", false }
func NewBusySource(mediaDebounce time.Duration) *BusySource { return &BusySource{} }
func (b *BusySource) SetMediaDebounce(d time.Duration)      {}

type OverlayAction int

const (
	OverlaySkip     OverlayAction = 1
	OverlayPostpone OverlayAction = 2
)

type OverlayOptions struct {
	Kind, Title, Timer, Tip, HexAccent      string
	ShowActions, Fullscreen, CurrentScreenOnly bool
}

func ScreenCount() int                            { return 1 }
func BringMainWindowForward()                     {}
func SetAccessoryActivationPolicy()               {}
func SetOverlayActionHandler(func(OverlayAction)) {}
func CreateOverlays(opts OverlayOptions)          {}
func UpdateOverlayTimer(timer string)             {}
func CloseOverlays()                              {}
