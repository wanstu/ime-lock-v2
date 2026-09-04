package main

import (
	"sync"

	"github.com/gogpu/systray"
)

type TrayManager struct {
	app  *App
	icon []byte

	mu       sync.Mutex
	started  bool
	stopping bool
	stopFunc func()
}

func NewTrayManager(app *App, icon []byte) *TrayManager {
	return &TrayManager{app: app, icon: icon}
}

func (t *TrayManager) Start() {
	if t == nil || t.app == nil {
		return
	}
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return
	}
	t.started = true
	t.stopping = false
	t.mu.Unlock()

	go func() {
		tray := systray.New()
		menu := systray.NewMenu()
		menu.Add("打开主面板", func() { t.app.showMainWindow() })
		menu.AddSeparator()
		menu.Add("切换自动修复", func() {
			if state, err := t.app.GetState(); err == nil {
				_, _ = t.app.SetAutoFix(!state.AutoFix)
			}
		})
		menu.Add("切换开机启动", func() {
			if state, err := t.app.GetState(); err == nil {
				_, _ = t.app.SetAutoStart(!state.AutoStart)
			}
		})
		menu.Add("切换静默启动", func() {
			if state, err := t.app.GetState(); err == nil {
				_, _ = t.app.SetSilentStart(!state.SilentStart)
			}
		})
		menu.Add("切换日志采集", func() {
			if state, err := t.app.GetState(); err == nil {
				_, _ = t.app.SetCaptureLogs(!state.CaptureLogs)
			}
		})
		menu.AddSeparator()
		menu.Add("退出", func() { t.app.quitApplication() })

		tray.SetIcon(t.icon).
			SetTooltip("IME Lock v2").
			SetMenu(menu)
		tray.OnClick(func() { t.app.showMainWindow() })
		tray.Show()

		t.mu.Lock()
		if t.stopping {
			t.mu.Unlock()
			tray.Remove()
			return
		}
		t.stopFunc = func() { tray.Remove() }
		t.mu.Unlock()

		_ = tray.Run()

		t.mu.Lock()
		t.started = false
		t.stopFunc = nil
		t.mu.Unlock()
	}()
}

func (t *TrayManager) Stop() {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.stopping = true
	stop := t.stopFunc
	t.mu.Unlock()
	if stop != nil {
		stop()
	}
}
