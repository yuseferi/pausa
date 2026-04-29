// Command pausa is the macOS break-reminder application. It is a thin
// entry-point that wires together the configuration store, scheduler,
// macOS bridge, and Wails runtime.
package main

import (
	"embed"
	"log/slog"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"pausa/internal/breakapp"
	"pausa/internal/clock"
	"pausa/internal/config"
	plog "pausa/internal/log"
	"pausa/internal/macos"
	"pausa/internal/scheduler"
	"pausa/internal/tips"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logFile := plog.Init(slog.LevelInfo)
	if logFile != nil {
		defer logFile.Close()
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		slog.Error("locate config", "err", err)
		return
	}
	cfgStore, err := config.Open(cfgPath)
	if err != nil {
		// Non-fatal; defaults are in use.
		slog.Warn("open config", "err", err, "path", cfgPath)
	}

	cfgSub := cfgStore.Subscribe()
	catalog := tips.NewCatalog()
	busySrc := macos.NewBusySource(cfgStore.Get().Idle.BusyMediaDebounce.AsDuration())
	sched := scheduler.New(
		clock.New(),
		cfgStore.Get(),
		cfgSub,
		macos.IdleSource{},
		busySrc,
	)
	app := breakapp.New(cfgStore, sched, catalog, cfgSub, busySrc)

	appMenu := buildAppMenu(app)

	err = wails.Run(&options.App{
		Title:             "Pausa",
		Width:             640,
		Height:            720,
		MinWidth:          560,
		MinHeight:         640,
		HideWindowOnClose: true,
		StartHidden:       false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:             appMenu,
		BackgroundColour: &options.RGBA{R: 17, G: 24, B: 39, A: 1},
		OnStartup:        app.Startup,
		OnDomReady:       app.DomReady,
		OnShutdown:       app.Shutdown,
		Bind:             []interface{}{app},
		Frameless:        true,
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  true,
				HideTitleBar:               true,
				FullSizeContent:            true,
			},
			About: &mac.AboutInfo{
				Title:   "Pausa",
				Message: "A mindful break reminder for macOS.",
			},
		},
	})
	if err != nil {
		slog.Error("wails run", "err", err)
	}
}

// buildAppMenu constructs the macOS menu bar (the one at the top of the
// screen, not the status item). Most operations live in the status bar; this
// menu provides standard shortcuts.
func buildAppMenu(app *breakapp.App) *menu.Menu {
	m := menu.NewMenu()
	m.Append(menu.AppMenu())

	file := m.AddSubmenu("File")
	file.AddText("Preferences…", keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
		app.OpenPreferences()
	})
	file.AddSeparator()
	file.AddText("Quit Pausa", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		app.QuitApp()
	})

	br := m.AddSubmenu("Breaks")
	br.AddText("Take a Break Now", keys.CmdOrCtrl("b"), func(_ *menu.CallbackData) {
		app.TakeBreakNow()
	})
	br.AddText("Skip Break", keys.Combo("s", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		app.SkipBreak()
	})
	br.AddText("Postpone Break", keys.Combo("p", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		app.PostponeBreak()
	})
	br.AddSeparator()
	br.AddText("Pause Breaks", keys.CmdOrCtrl("p"), func(_ *menu.CallbackData) {
		app.PauseBreaks()
	})
	br.AddText("Resume Breaks", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		app.ResumeBreaks()
	})
	br.AddSeparator()
	br.AddText("Reset Schedule", nil, func(_ *menu.CallbackData) {
		app.ResetBreaks()
	})

	m.Append(menu.EditMenu())
	return m
}
