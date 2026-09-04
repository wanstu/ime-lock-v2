//go:build windows

package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	ole "github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

const (
	eventObjectNameChange = 0x800C
	wineventOutOfContext  = 0x0000

	wmHotkey = 0x0312
	wmQuit   = 0x0012

	modControl  = 0x0002
	modShift    = 0x0004
	modNoRepeat = 0x4000
	vkF9        = 0x78
	hotkeyID    = 1

	wmIMEControl     = 0x0283
	imcSetOpenStatus = 0x0002

	dispidAccName = -5003
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	imm32  = windows.NewLazySystemDLL("imm32.dll")
	oleacc = windows.NewLazySystemDLL("oleacc.dll")

	procSetWinEventHook           = user32.NewProc("SetWinEventHook")
	procUnhookWinEvent            = user32.NewProc("UnhookWinEvent")
	procGetMessageW               = user32.NewProc("GetMessageW")
	procTranslateMessage          = user32.NewProc("TranslateMessage")
	procDispatchMessageW          = user32.NewProc("DispatchMessageW")
	procPeekMessageW              = user32.NewProc("PeekMessageW")
	procPostThreadMessageW        = user32.NewProc("PostThreadMessageW")
	procRegisterHotKey            = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey          = user32.NewProc("UnregisterHotKey")
	procGetForegroundWindow       = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW            = user32.NewProc("GetWindowTextW")
	procSendMessageW              = user32.NewProc("SendMessageW")
	procImmGetDefaultIMEWnd       = imm32.NewProc("ImmGetDefaultIMEWnd")
	procAccessibleObjectFromEvent = oleacc.NewProc("AccessibleObjectFromEvent")

	activeIMEWatcher atomic.Pointer[IMEWatcher]
	winEventCallback = windows.NewCallback(winEventProc)
)

type winPoint struct {
	X int32
	Y int32
}

type winMsg struct {
	HWnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       winPoint
	LPrivate uint32
}

type IMEWatcher struct {
	enabled atomic.Bool
	running atomic.Bool
	pending atomic.Bool
	thread  atomic.Uint32

	mu   sync.Mutex
	done chan struct{}

	onLog    func(string)
	onFix    func()
	onToggle func(bool)
}

func NewIMEWatcher(onLog func(string), onFix func(), onToggle func(bool)) *IMEWatcher {
	return &IMEWatcher{onLog: onLog, onFix: onFix, onToggle: onToggle}
}

func (w *IMEWatcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running.Load() {
		return nil
	}

	for _, proc := range []*windows.LazyProc{
		procSetWinEventHook,
		procUnhookWinEvent,
		procGetMessageW,
		procPostThreadMessageW,
		procRegisterHotKey,
		procUnregisterHotKey,
		procGetForegroundWindow,
		procGetWindowTextW,
		procSendMessageW,
		procImmGetDefaultIMEWnd,
		procAccessibleObjectFromEvent,
	} {
		if err := proc.Find(); err != nil {
			return fmt.Errorf("初始化 Windows 输入法监听 API 失败: %w", err)
		}
	}

	done := make(chan struct{})
	initialized := make(chan error, 1)
	w.done = done
	go w.runMessageLoop(done, initialized)

	if err := <-initialized; err != nil {
		w.done = nil
		return err
	}
	w.running.Store(true)
	return nil
}

func (w *IMEWatcher) Stop() {
	w.mu.Lock()
	if w.done == nil {
		w.mu.Unlock()
		return
	}
	done := w.done
	threadID := w.thread.Load()
	w.done = nil
	w.running.Store(false)
	w.mu.Unlock()

	if threadID != 0 {
		_, _, _ = procPostThreadMessageW.Call(uintptr(threadID), wmQuit, 0, 0)
	}
	<-done
}

func (w *IMEWatcher) SetEnabled(enabled bool) {
	w.enabled.Store(enabled)
	if !enabled {
		w.pending.Store(false)
	}
}

func (w *IMEWatcher) Enabled() bool {
	return w != nil && w.enabled.Load()
}

func (w *IMEWatcher) Running() bool {
	return w != nil && w.running.Load()
}

func (w *IMEWatcher) runMessageLoop(done chan<- struct{}, initialized chan<- error) {
	defer close(done)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		initialized <- fmt.Errorf("初始化 COM 失败: %w", err)
		return
	}
	defer ole.CoUninitialize()

	threadID := windows.GetCurrentThreadId()
	w.thread.Store(threadID)
	defer w.thread.Store(0)

	// 强制创建线程消息队列，确保 Stop() 的 PostThreadMessageW(WM_QUIT) 可用。
	var msg winMsg
	_, _, _ = procPeekMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, 0)

	activeIMEWatcher.Store(w)
	defer activeIMEWatcher.CompareAndSwap(w, nil)

	hook, _, hookErr := procSetWinEventHook.Call(
		eventObjectNameChange,
		eventObjectNameChange,
		0,
		winEventCallback,
		0,
		0,
		wineventOutOfContext,
	)
	if hook == 0 {
		initialized <- fmt.Errorf("SetWinEventHook(EVENT_OBJECT_NAMECHANGE) 失败: %v", hookErr)
		return
	}
	defer procUnhookWinEvent.Call(hook)

	hotkeyOK, _, hotkeyErr := procRegisterHotKey.Call(
		0,
		hotkeyID,
		modControl|modShift|modNoRepeat,
		vkF9,
	)
	if hotkeyOK != 0 {
		defer procUnregisterHotKey.Call(0, hotkeyID)
	} else if w.onLog != nil {
		w.onLog(fmt.Sprintf("注册快捷键 Ctrl+Shift+F9 失败: %v", hotkeyErr))
	}

	initialized <- nil

	for {
		result, _, getErr := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) == -1 {
			if w.onLog != nil {
				w.onLog(fmt.Sprintf("Windows 消息循环异常: %v", getErr))
			}
			return
		}
		if result == 0 || msg.Message == wmQuit {
			return
		}
		if msg.Message == wmHotkey && msg.WParam == hotkeyID {
			next := !w.enabled.Load()
			w.SetEnabled(next)
			if w.onToggle != nil {
				w.onToggle(next)
			}
			continue
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		_, _, _ = procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func winEventProc(
	_ uintptr,
	event uint32,
	hwnd uintptr,
	objectID int32,
	childID int32,
	_ uint32,
	_ uint32,
) uintptr {
	if event != eventObjectNameChange || hwnd == 0 {
		return 0
	}
	w := activeIMEWatcher.Load()
	if w == nil {
		return 0
	}

	name := accessibleNameFromEvent(hwnd, objectID, childID)
	switch inputIndicatorMode(name) {
	case "english":
		w.detectEnglishMode(name)
	case "chinese":
		w.pending.Store(false)
		if w.onLog != nil {
			w.onLog(fmt.Sprintf("[DETECT] accName=%q -> 中文模式（正常，不动作）", name))
		}
	}
	return 0
}

func inputIndicatorMode(name string) string {
	if !strings.Contains(name, "任务栏输入指示") {
		return ""
	}
	if strings.Contains(name, "英语模式") {
		return "english"
	}
	if strings.Contains(name, "中文模式") {
		return "chinese"
	}
	return ""
}

func (w *IMEWatcher) detectEnglishMode(accName string) {
	if w.onLog != nil {
		w.onLog(fmt.Sprintf("[DETECT] accName=%q -> 英文子模式%s", accName, foregroundInfo()))
	}
	if !w.enabled.Load() {
		if w.onLog != nil {
			w.onLog("[SKIP] 自动修复已关闭，不动作")
		}
		return
	}
	if !w.pending.CompareAndSwap(false, true) {
		if w.onLog != nil {
			w.onLog("[DEBOUNCE] 已发送修复，等待中文模式确认")
		}
		return
	}

	if w.onLog != nil {
		w.onLog("[FIX] 发送 WM_IME_CONTROL 打开 IME（中文）")
	}
	if setIMEChinese() {
		if w.onFix != nil {
			w.onFix()
		}
	} else {
		w.pending.Store(false)
		if w.onLog != nil {
			w.onLog("[FIX] 未找到前台窗口或默认 IME 窗口，未发送修复")
		}
	}
}

func accessibleNameFromEvent(hwnd uintptr, objectID, childID int32) string {
	var dispatch *ole.IDispatch
	var child ole.VARIANT

	hr, _, _ := procAccessibleObjectFromEvent.Call(
		hwnd,
		uintptr(uint32(objectID)),
		uintptr(uint32(childID)),
		uintptr(unsafe.Pointer(&dispatch)),
		uintptr(unsafe.Pointer(&child)),
	)
	if int32(hr) < 0 || dispatch == nil {
		return ""
	}
	defer dispatch.Release()
	defer child.Clear()

	result, err := dispatch.Invoke(dispidAccName, ole.DISPATCH_PROPERTYGET, int32(child.Val))
	if err != nil || result == nil {
		return ""
	}
	defer result.Clear()
	return strings.TrimSpace(result.ToString())
}

func foregroundInfo() string {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}
	title := foregroundTitle(hwnd)
	if title == "" {
		return ""
	}
	return fmt.Sprintf(" (fg=%q)", title)
}

func foregroundTitle(hwnd uintptr) string {
	var buf [256]uint16
	length, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if length == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:length])
}

func setIMEChinese() bool {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}
	imeWnd, _, _ := procImmGetDefaultIMEWnd.Call(hwnd)
	if imeWnd == 0 {
		return false
	}
	_, _, _ = procSendMessageW.Call(imeWnd, wmIMEControl, imcSetOpenStatus, 1)
	return true
}
