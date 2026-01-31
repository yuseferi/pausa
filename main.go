package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create menu for the menu bar
	appMenu := menu.NewMenu()

	// App Menu (Pausa)
	appMenu.Append(menu.AppMenu())

	// File Menu
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Preferences...", keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "openPreferences", nil)
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit Pausa", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})

	// Breaks Menu
	breaksMenu := appMenu.AddSubmenu("Breaks")
	breaksMenu.AddText("Take a Break Now", keys.CmdOrCtrl("b"), func(_ *menu.CallbackData) {
		app.TakeBreakNow()
	})
	breaksMenu.AddSeparator()
	breaksMenu.AddText("Pause Breaks", keys.CmdOrCtrl("p"), func(_ *menu.CallbackData) {
		app.PauseBreaks()
	})
	breaksMenu.AddText("Resume Breaks", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		app.ResumeBreaks()
	})
	breaksMenu.AddSeparator()
	breaksMenu.AddText("Reset Breaks", nil, func(_ *menu.CallbackData) {
		app.ResetBreaks()
	})

	// Help Menu
	appMenu.Append(menu.EditMenu())

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "Pausa",
		Width:            400,
		Height:           300,
		MinWidth:         400,
		MinHeight:        300,
		DisableResize:    false,
		Fullscreen:       false,
		StartHidden:      false,
		HideWindowOnClose: true,
		AlwaysOnTop:      false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:             appMenu,
		BackgroundColour: &options.RGBA{R: 45, G: 45, B: 45, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnDomReady:       app.domReady,
		Bind: []interface{}{
			app,
		},
		Frameless: true,
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  true,
				HideTitleBar:               true,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			WindowIsTranslucent: false,
			About: &mac.AboutInfo{
				Title:   "Pausa",
				Message: "The break time reminder app\n\nTake regular breaks to stay healthy and productive.",
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}