package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const logCapacity = 600

type AppState struct {
	Running     bool   `json:"running"`
	AutoFix     bool   `json:"auto_fix"`
	AutoStart   bool   `json:"auto_start"`
	SilentStart bool   `json:"silent_start"`
	CaptureLogs bool   `json:"capture_logs"`
	FixCount    uint64 `json:"fix_count"`
	LastFixAt   string `json:"last_fix_at"`
	LastError   string `json:"last_error"`
	Shortcut    string `json:"shortcut"`
	ConfigPath  string `json:"config_path"`
}

type App struct {
	ctx context.Context

	store   *ConfigStore
	config  Config
	watcher *IMEWatcher
	tray    *TrayManager
	initErr error

	mu          sync.RWMutex
	captureLogs bool
	logs        []string
	fixCount    uint64
	lastFixAt   time.Time
	lastError   string
}

func NewApp() *App {
	store, err := NewDefaultConfigStore()
	app := &App{store: store, initErr: err}
	if err == nil {
		cfg, loadErr := store.Load()
		if loadErr != nil {
			app.initErr = loadErr
		} else {
			app.config = cfg
		}
	}

	app.watcher = NewIMEWatcher(app.appendLog, app.recordFix, app.handleHotkeyToggle)
	app.watcher.SetEnabled(true)
	return app
}

func (a *App) attachTray(icon []byte) {
	if a != nil {
		a.tray = NewTrayManager(a, icon)
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go watchSingleInstanceWake(ctx, a.showMainWindow)
	if a.watcher != nil {
		if err := a.watcher.Start(); err != nil {
			a.setError(err)
		}
	}
}

func (a *App) domReady(ctx context.Context) {
	a.ctx = ctx
	if a.tray != nil {
		a.tray.Start()
	}
}

func (a *App) shutdown(context.Context) {
	if a.watcher != nil {
		a.watcher.Stop()
	}
	if a.tray != nil {
		a.tray.Stop()
	}
}

func (a *App) SilentStart() bool {
	if a == nil {
		return false
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config.SilentStart
}

func (a *App) GetState() (AppState, error) {
	if err := a.ready(); err != nil {
		return AppState{}, err
	}
	a.mu.RLock()
	defer a.mu.RUnlock()

	configPath := ""
	if a.store != nil {
		configPath = a.store.path
	}
	lastFix := ""
	if !a.lastFixAt.IsZero() {
		lastFix = a.lastFixAt.Format("2006-01-02 15:04:05")
	}
	return AppState{
		Running:     a.watcher.Running(),
		AutoFix:     a.watcher.Enabled(),
		AutoStart:   a.config.AutoStart,
		SilentStart: a.config.SilentStart,
		CaptureLogs: a.captureLogs,
		FixCount:    a.fixCount,
		LastFixAt:   lastFix,
		LastError:   a.lastError,
		Shortcut:    "Ctrl + Shift + F9",
		ConfigPath:  configPath,
	}, nil
}

func (a *App) SetAutoFix(enabled bool) (AppState, error) {
	if err := a.ready(); err != nil {
		return AppState{}, err
	}
	a.watcher.SetEnabled(enabled)
	a.appendLog(fmt.Sprintf("自动修复已%s", onOff(enabled)))
	return a.GetState()
}

func (a *App) SetAutoStart(enabled bool) (AppState, error) {
	if err := a.ready(); err != nil {
		return AppState{}, err
	}
	if err := setManagerAutoStart(enabled); err != nil {
		a.setError(err)
		return AppState{}, err
	}
	a.mu.Lock()
	a.config.AutoStart = enabled
	cfg := a.config
	a.mu.Unlock()
	if err := a.store.Save(cfg); err != nil {
		a.setError(err)
		return AppState{}, err
	}
	a.appendLog(fmt.Sprintf("开机启动已%s", onOff(enabled)))
	return a.GetState()
}

func (a *App) SetSilentStart(enabled bool) (AppState, error) {
	if err := a.ready(); err != nil {
		return AppState{}, err
	}
	a.mu.Lock()
	a.config.SilentStart = enabled
	cfg := a.config
	a.mu.Unlock()
	if err := a.store.Save(cfg); err != nil {
		a.setError(err)
		return AppState{}, err
	}
	a.appendLog(fmt.Sprintf("静默启动已%s", onOff(enabled)))
	return a.GetState()
}

func (a *App) SetCaptureLogs(enabled bool) (AppState, error) {
	if err := a.ready(); err != nil {
		return AppState{}, err
	}
	a.mu.Lock()
	a.captureLogs = enabled
	a.mu.Unlock()
	if enabled {
		a.appendLog("日志采集已开始")
	}
	return a.GetState()
}

func (a *App) GetLogs() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return strings.Join(a.logs, "\n"), nil
}

func (a *App) ClearLogs() error {
	if err := a.ready(); err != nil {
		return err
	}
	a.mu.Lock()
	a.logs = nil
	a.mu.Unlock()
	return nil
}

func (a *App) handleHotkeyToggle(enabled bool) {
	a.appendLog(fmt.Sprintf("快捷键切换：自动修复已%s", onOff(enabled)))
}

func (a *App) recordFix() {
	a.mu.Lock()
	a.fixCount++
	a.lastFixAt = time.Now()
	a.mu.Unlock()
}

func (a *App) appendLog(message string) {
	if a == nil || strings.TrimSpace(message) == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.captureLogs {
		return
	}
	line := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), message)
	a.logs = append(a.logs, line)
	if len(a.logs) > logCapacity {
		a.logs = append([]string(nil), a.logs[len(a.logs)-logCapacity:]...)
	}
}

func (a *App) setError(err error) {
	if a == nil || err == nil {
		return
	}
	a.mu.Lock()
	a.lastError = err.Error()
	a.mu.Unlock()
}

func (a *App) showMainWindow() {
	if a == nil || a.ctx == nil {
		return
	}
	runtime.WindowUnminimise(a.ctx)
	runtime.Show(a.ctx)
}

func (a *App) quitApplication() {
	if a == nil || a.ctx == nil {
		return
	}
	runtime.Quit(a.ctx)
}

func (a *App) ready() error {
	if a == nil {
		return errors.New("应用未初始化")
	}
	if a.initErr != nil {
		return a.initErr
	}
	if a.store == nil || a.watcher == nil {
		return errors.New("核心服务未初始化")
	}
	return nil
}

func onOff(v bool) string {
	if v {
		return "开启"
	}
	return "关闭"
}
