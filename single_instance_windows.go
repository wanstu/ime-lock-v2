//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

const (
	instanceLockFileName = "app.lock"
	instanceWakeFileName = "show-window.request"
)

func acquireSingleInstance() (func(), bool, error) {
	dir, err := configDir()
	if err != nil {
		return nil, false, err
	}
	return acquireInstanceLock(filepath.Join(dir, instanceLockFileName))
}

func acquireInstanceLock(lockPath string) (func(), bool, error) {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, false, fmt.Errorf("无法创建单实例锁目录: %w", err)
	}
	pathPtr, err := windows.UTF16PtrFromString(lockPath)
	if err != nil {
		return nil, false, fmt.Errorf("单实例锁路径无效: %w", err)
	}
	handle, err := windows.CreateFile(
		pathPtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		if errors.Is(err, windows.ERROR_SHARING_VIOLATION) || errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return func() {}, false, nil
		}
		return nil, false, fmt.Errorf("获取 IME Lock v2 单实例锁失败: %w", err)
	}
	released := false
	return func() {
		if released {
			return
		}
		released = true
		_ = windows.CloseHandle(handle)
	}, true, nil
}

func singleInstanceWakePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, instanceWakeFileName), nil
}

func prepareSingleInstanceWake() error {
	path, err := singleInstanceWakePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("清理单实例唤醒请求失败: %w", err)
	}
	return nil
}

func requestExistingInstanceWindow() error {
	path, err := singleInstanceWakePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("创建单实例唤醒目录失败: %w", err)
	}
	if err := os.WriteFile(path, []byte(time.Now().Format(time.RFC3339Nano)), 0o600); err != nil {
		return fmt.Errorf("发送单实例唤醒请求失败: %w", err)
	}
	return nil
}

func consumeSingleInstanceWake() (bool, error) {
	path, err := singleInstanceWakePath()
	if err != nil {
		return false, err
	}
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("读取单实例唤醒请求失败: %w", err)
	}
	return true, nil
}

func watchSingleInstanceWake(ctx context.Context, onRequest func()) {
	if onRequest == nil {
		return
	}
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			requested, err := consumeSingleInstanceWake()
			if err == nil && requested {
				onRequest()
			}
		}
	}
}
