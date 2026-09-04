package main

import (
	"embed"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/src
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	releaseInstance, primary, err := acquireSingleInstance()
	if err != nil {
		println("Error:", err.Error())
		return
	}
	if !primary {
		if !launchedFromAutoStart() {
			_ = requestExistingInstanceWindow()
		}
		return
	}
	defer releaseInstance()
	_ = prepareSingleInstanceWake()

	app := NewApp()
	app.attachTray(appIcon)

	err = wails.Run(&options.App{
		Title:             "IME Lock v2",
		Width:             760,
		Height:            680,
		MinWidth:          680,
		MinHeight:         560,
		StartHidden:       launchedFromAutoStart() && app.SilentStart(),
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 244, G: 247, B: 251, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}

func launchedFromAutoStart() bool {
	return hasAutoStartArg(os.Args[1:])
}

func hasAutoStartArg(args []string) bool {
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if strings.EqualFold(arg, "--autostart") || strings.EqualFold(arg, "--minimized") {
			return true
		}
	}
	return false
}
