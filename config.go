package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wanstu/wails-desktop-kit/jsonstore"
	kitpaths "github.com/wanstu/wails-desktop-kit/paths"
	kittheme "github.com/wanstu/wails-desktop-kit/theme"
)

const (
	themeLight       = "light"
	themeDark        = "dark"
	themeSystem      = "system"
	defaultThemePack = "aurora"
)

type Config struct {
	AutoStart   bool   `json:"auto_start"`
	SilentStart bool   `json:"silent_start"`
	Theme       string `json:"theme"`
	ThemePack   string `json:"theme_pack"`
}

type ConfigStore struct {
	path string
}

func defaultConfig() Config {
	return Config{Theme: themeLight, ThemePack: defaultThemePack}
}

func normalizeTheme(theme string) string {
	theme = strings.ToLower(strings.TrimSpace(theme))
	if err := kittheme.ValidateMode(kittheme.Mode(theme)); err == nil {
		return theme
	}
	return themeLight
}

func validTheme(theme string) bool {
	return kittheme.ValidateMode(kittheme.Mode(theme)) == nil
}

func normalizeThemePack(pack string) string {
	pack = strings.ToLower(strings.TrimSpace(pack))
	if validThemePack(pack) {
		return pack
	}
	return defaultThemePack
}

func validThemePack(pack string) bool {
	return kittheme.ValidatePackName(pack) == nil
}

func NewDefaultConfigStore() (*ConfigStore, error) {
	dir, err := kitpaths.ConfigDir("ime-lock-v2")
	if err != nil {
		return nil, fmt.Errorf("无法获取用户目录: %w", err)
	}
	legacyDir, err := kitpaths.ConfigDir("img-lock-v2")
	if err != nil {
		return nil, fmt.Errorf("无法获取旧配置目录: %w", err)
	}
	path := filepath.Join(dir, "config.json")
	legacyPath := filepath.Join(legacyDir, "config.json")
	if err := migrateConfigFile(legacyPath, path); err != nil {
		return nil, err
	}
	return &ConfigStore{path: path}, nil
}

func configDir() (string, error) {
	dir, err := kitpaths.ConfigDir("ime-lock-v2")
	if err != nil {
		return "", fmt.Errorf("无法获取用户目录: %w", err)
	}
	return dir, nil
}

func migrateConfigFile(legacyPath, newPath string) error {
	if _, err := kitpaths.MigrateFileIfMissing(legacyPath, newPath); err != nil {
		return fmt.Errorf("迁移旧配置失败: %w", err)
	}
	return nil
}

func (s *ConfigStore) values() *jsonstore.Store[Config] {
	return jsonstore.New(s.path, jsonstore.Options[Config]{
		Default: defaultConfig,
		Normalize: func(cfg *Config) {
			cfg.Theme = normalizeTheme(cfg.Theme)
			cfg.ThemePack = normalizeThemePack(cfg.ThemePack)
		},
	})
}

func (s *ConfigStore) Load() (Config, error) {
	if s == nil || s.path == "" {
		return Config{}, errors.New("配置存储未初始化")
	}
	cfg, err := s.values().Load()
	if err != nil {
		return Config{}, fmt.Errorf("读取配置失败: %w", err)
	}
	return cfg, nil
}

func (s *ConfigStore) Save(cfg Config) error {
	if s == nil || s.path == "" {
		return errors.New("配置存储未初始化")
	}
	if err := s.values().Save(cfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	return nil
}
