package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"

	desktopkit "github.com/wanstu/wails-desktop-kit"
	"github.com/wanstu/wails-desktop-kit/autostart"
	kitui "github.com/wanstu/wails-desktop-kit/ui"
)

const (
	desktopAppID = "com.wanstu.ime-lock-v2"
	autoStartID  = "IME-Lock-V2"
)

//go:embed all:frontend/src
var assets embed.FS

//go:embed assets/icons/tray-icon.png
var trayIcon []byte

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

func run() error {
	appAssets, err := fs.Sub(assets, "frontend/src")
	if err != nil {
		return fmt.Errorf("加载前端资源失败: %w", err)
	}

	login, err := autostart.New(autostart.Config{
		ID:          autoStartID,
		DisplayName: "IME Lock v2",
		Comment:     "Windows 输入法状态守护",
		Arguments:   []string{"--autostart"},
	})
	if err != nil {
		return err
	}

	app := NewApp(login)
	launch := desktopkit.LaunchOptions{AutoStart: launchedFromAutoStart()}
	window := desktopkit.DefaultWindowConfig()
	window.Width = 760
	window.Height = 680
	window.MinWidth = 680
	window.MinHeight = 560
	window.StartHiddenOnAutoStart = app.SilentStart()
	window.Background = desktopkit.Color{R: 244, G: 247, B: 251, A: 1}

	return desktopkit.Run(desktopkit.Config{
		ID:             desktopAppID,
		Title:          "IME Lock v2",
		Assets:         kitui.Mount(appAssets),
		Bind:           []interface{}{app},
		Launch:         launch,
		Window:         window,
		Tray:           imeTrayConfig(app, trayIcon),
		Theme:          desktopkit.DefaultThemeConfig(),
		SingleInstance: true,
		SecondInstance: handleSecondInstance,
		Hooks: desktopkit.Hooks{
			Startup:  app.startup,
			Shutdown: app.shutdown,
		},
	})
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
