package main

import (
	"os"
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

func TestMigrateConfigFile(t *testing.T) {
	root := t.TempDir()
	legacyPath := filepath.Join(root, "img-lock-v2", "config.json")
	newPath := filepath.Join(root, "ime-lock-v2", "config.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	want := []byte("{\"auto_start\":true,\"silent_start\":true}\n")
	if err := os.WriteFile(legacyPath, want, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := migrateConfigFile(legacyPath, newPath); err != nil {
		t.Fatalf("migrateConfigFile() error = %v", err)
	}
	got, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("migrated config = %q, want %q", got, want)
	}
}
