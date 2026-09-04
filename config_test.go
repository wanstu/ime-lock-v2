package main

import (
	"path/filepath"
	"testing"
)

func TestConfigStoreRoundTrip(t *testing.T) {
	store := &ConfigStore{path: filepath.Join(t.TempDir(), "config.json")}
	want := Config{AutoStart: true, SilentStart: true}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}

	updated := Config{AutoStart: false, SilentStart: true}
	if err := store.Save(updated); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}
	got, err = store.Load()
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}
	if got != updated {
		t.Fatalf("second Load() = %+v, want %+v", got, updated)
	}
}

func TestConfigStoreMissingFileUsesDefaults(t *testing.T) {
	store := &ConfigStore{path: filepath.Join(t.TempDir(), "missing.json")}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != (Config{}) {
		t.Fatalf("Load() = %+v, want zero config", got)
	}
}
