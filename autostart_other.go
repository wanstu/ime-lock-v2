//go:build !windows

package main

import "errors"

func managerAutoStartEnabled() (bool, error) {
	return false, nil
}

func setManagerAutoStart(bool) error {
	return errors.New("开机启动目前仅支持 Windows")
}

func managerAutoStartCommand() (string, error) {
	return "", errors.New("开机启动目前仅支持 Windows")
}
