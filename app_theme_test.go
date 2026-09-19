package main

import (
	"path/filepath"
	"testing"

	"github.com/wanstu/wails-desktop-kit/autostart"
)

func newThemeTestApp(t *testing.T) *App {
	t.Helper()

	manager, err := autostart.New(autostart.Config{
		ID:             "IME-Lock-V2-Test",
		DisplayName:    "IME Lock v2 Test",
		ExecutablePath: filepath.Join(t.TempDir(), "ime-lock-v2.exe"),
		Arguments:      []string{"--autostart"},
	})
	if err != nil {
		t.Fatalf("autostart.New() error = %v", err)
	}

	store := &ConfigStore{path: filepath.Join(t.TempDir(), "config.json")}
	app := &App{
		store:     store,
		config:    defaultConfig(),
		watcher:   NewIMEWatcher(nil, nil, nil),
		autoStart: manager,
	}
	return app
}

func TestSetThemePersistsAndReturnsState(t *testing.T) {
	app := newThemeTestApp(t)

	state, err := app.SetTheme(themeDark)
	if err != nil {
		t.Fatalf("SetTheme() error = %v", err)
	}
	if state.Theme != themeDark {
		t.Fatalf("state.Theme = %q, want %q", state.Theme, themeDark)
	}

	saved, err := app.store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if saved.Theme != themeDark {
		t.Fatalf("saved.Theme = %q, want %q", saved.Theme, themeDark)
	}

	state, err = app.SetTheme(themeSystem)
	if err != nil {
		t.Fatalf("SetTheme(system) error = %v", err)
	}
	if state.Theme != themeSystem {
		t.Fatalf("state.Theme = %q, want %q", state.Theme, themeSystem)
	}
}

func TestSetThemePackPersistsFutureRuntimeTheme(t *testing.T) {
	app := newThemeTestApp(t)

	state, err := app.SetThemePack("future-theme")
	if err != nil {
		t.Fatalf("SetThemePack() error = %v", err)
	}
	if state.ThemePack != "future-theme" {
		t.Fatalf("state.ThemePack = %q, want %q", state.ThemePack, "future-theme")
	}

	saved, err := app.store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if saved.ThemePack != "future-theme" {
		t.Fatalf("saved.ThemePack = %q, want %q", saved.ThemePack, "future-theme")
	}
}

func TestSetThemePackRejectsUnsafeValueWithoutChangingConfig(t *testing.T) {
	app := newThemeTestApp(t)

	if _, err := app.SetThemePack("../evil"); err == nil {
		t.Fatal("SetThemePack(../evil) expected error")
	}
	if app.config.ThemePack != defaultThemePack {
		t.Fatalf("config.ThemePack = %q, want %q", app.config.ThemePack, defaultThemePack)
	}
}

func TestSetThemeRejectsUnknownValueWithoutChangingConfig(t *testing.T) {
	app := newThemeTestApp(t)

	if _, err := app.SetTheme("midnight"); err == nil {
		t.Fatal("SetTheme(midnight) expected error")
	}
	if app.config.Theme != themeLight {
		t.Fatalf("config.Theme = %q, want %q", app.config.Theme, themeLight)
	}
}
