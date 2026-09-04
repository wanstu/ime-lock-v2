package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	AutoStart   bool `json:"auto_start"`
	SilentStart bool `json:"silent_start"`
}

type ConfigStore struct {
	path string
}

func NewDefaultConfigStore() (*ConfigStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户目录: %w", err)
	}
	path := filepath.Join(home, ".config", "ime-lock-v2", "config.json")
	legacyPath := filepath.Join(home, ".config", "img-lock-v2", "config.json")
	if err := migrateConfigFile(legacyPath, path); err != nil {
		return nil, err
	}
	return &ConfigStore{path: path}, nil
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户目录: %w", err)
	}
	return filepath.Join(home, ".config", "ime-lock-v2"), nil
}

func migrateConfigFile(legacyPath, newPath string) error {
	if _, err := os.Stat(newPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("检查新配置失败: %w", err)
	}

	data, err := os.ReadFile(legacyPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取旧配置失败: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return fmt.Errorf("创建新配置目录失败: %w", err)
	}
	if err := os.WriteFile(newPath, data, 0o600); err != nil {
		return fmt.Errorf("迁移旧配置失败: %w", err)
	}
	return nil
}

func (s *ConfigStore) Load() (Config, error) {
	if s == nil || s.path == "" {
		return Config{}, errors.New("配置存储未初始化")
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("读取配置失败: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置失败: %w", err)
	}
	return cfg, nil
}

func (s *ConfigStore) Save(cfg Config) error {
	if s == nil || s.path == "" {
		return errors.New("配置存储未初始化")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("保存配置失败: %w", err)
	}
	return nil
}
