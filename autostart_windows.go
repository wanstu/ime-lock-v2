//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	managerRunKeyPath     = `Software\Microsoft\Windows\CurrentVersion\Run`
	managerRunValue       = "IME-Lock-V2"
	legacyManagerRunValue = "IMG-Lock-V2"
)

func managerAutoStartEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, managerRunKeyPath, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("读取开机启动配置失败: %w", err)
	}
	defer key.Close()

	expected, err := managerAutoStartCommand()
	if err != nil {
		return false, err
	}
	value, _, err := key.GetStringValue(managerRunValue)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("读取开机启动配置失败: %w", err)
	}
	return strings.EqualFold(strings.TrimSpace(value), expected), nil
}

func setManagerAutoStart(enabled bool) error {
	if !enabled {
		key, err := registry.OpenKey(registry.CURRENT_USER, managerRunKeyPath, registry.SET_VALUE)
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("打开开机启动配置失败: %w", err)
		}
		defer key.Close()
		if err := key.DeleteValue(managerRunValue); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return fmt.Errorf("关闭开机启动失败: %w", err)
		}
		if err := key.DeleteValue(legacyManagerRunValue); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return fmt.Errorf("清理旧开机启动项失败: %w", err)
		}
		return nil
	}

	command, err := managerAutoStartCommand()
	if err != nil {
		return err
	}
	key, _, err := registry.CreateKey(registry.CURRENT_USER, managerRunKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("创建开机启动配置失败: %w", err)
	}
	defer key.Close()
	if err := key.SetStringValue(managerRunValue, command); err != nil {
		return fmt.Errorf("保存开机启动配置失败: %w", err)
	}
	if err := key.DeleteValue(legacyManagerRunValue); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("清理旧开机启动项失败: %w", err)
	}
	return nil
}

func managerAutoStartCommand() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("无法获取程序路径: %w", err)
	}
	return fmt.Sprintf("\"%s\" --autostart", executable), nil
}
