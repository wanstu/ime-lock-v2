package main

import desktopkit "github.com/wanstu/wails-desktop-kit"

type appAutoStartProvider struct {
	app *App
}

func (p *appAutoStartProvider) Supported() bool {
	return p != nil && p.app != nil && p.app.autoStart != nil && p.app.autoStart.Supported()
}

func (p *appAutoStartProvider) Enabled() (bool, error) {
	if p == nil || p.app == nil {
		return false, nil
	}
	if err := p.app.ready(); err != nil {
		return false, err
	}
	p.app.mu.RLock()
	defer p.app.mu.RUnlock()
	return p.app.config.AutoStart, nil
}

func (p *appAutoStartProvider) SetEnabled(enabled bool) error {
	if p == nil || p.app == nil {
		return nil
	}
	return p.app.setAutoStart(enabled)
}

func imeTrayConfig(app *App, icon []byte) desktopkit.TrayConfig {
	return desktopkit.TrayConfig{
		Enabled:            true,
		Icon:               icon,
		Tooltip:            "IME Lock v2",
		ShowLabel:          "打开主面板",
		HideLabel:          "隐藏主面板",
		LaunchAtLoginLabel: "开机启动",
		QuitLabel:          "退出",
		Items: []desktopkit.TrayItem{
			stateCheckbox("自动修复", app, func(state AppState) bool { return state.AutoFix }, app.SetAutoFix),
			stateCheckbox("静默启动", app, func(state AppState) bool { return state.SilentStart }, app.SetSilentStart),
			stateCheckbox("采集日志", app, func(state AppState) bool { return state.CaptureLogs }, app.SetCaptureLogs),
		},
		AutoStart: &appAutoStartProvider{app: app},
	}
}

func stateCheckbox(
	label string,
	app *App,
	get func(AppState) bool,
	set func(bool) (AppState, error),
) desktopkit.TrayItem {
	return desktopkit.Checkbox(label, desktopkit.TrayCheckbox{
		Get: func(*desktopkit.Controller) (bool, error) {
			state, err := app.GetState()
			if err != nil {
				return false, err
			}
			return get(state), nil
		},
		Set: func(_ *desktopkit.Controller, enabled bool) error {
			_, err := set(enabled)
			return err
		},
	})
}
