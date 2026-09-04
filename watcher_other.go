//go:build !windows

package main

import (
	"errors"
	"sync/atomic"
)

type IMEWatcher struct {
	enabled atomic.Bool
	running atomic.Bool
}

func NewIMEWatcher(func(string), func(), func(bool)) *IMEWatcher {
	return &IMEWatcher{}
}

func (w *IMEWatcher) Start() error {
	return errors.New("IME 监听目前仅支持 Windows")
}

func (w *IMEWatcher) Stop() {
	w.running.Store(false)
}

func (w *IMEWatcher) SetEnabled(enabled bool) {
	w.enabled.Store(enabled)
}

func (w *IMEWatcher) Enabled() bool {
	return w != nil && w.enabled.Load()
}

func (w *IMEWatcher) Running() bool {
	return w != nil && w.running.Load()
}
