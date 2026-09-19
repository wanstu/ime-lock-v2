//go:build windows

package main

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const (
	legacyRunKeyPath  = `Software\Microsoft\Windows\CurrentVersion\Run`
	legacyRunValueKey = "IMG-Lock-V2"
)

func cleanupLegacyAutoStart() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, legacyRunKeyPath, registry.SET_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("打开旧开机启动配置失败: %w", err)
	}
	defer key.Close()
	if err := key.DeleteValue(legacyRunValueKey); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("清理旧开机启动项失败: %w", err)
	}
	return nil
}
