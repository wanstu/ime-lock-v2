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
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	return &ConfigStore{path: filepath.Join(dir, "config.json")}, nil
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户目录: %w", err)
	}
	return filepath.Join(home, ".config", "img-lock-v2"), nil
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
