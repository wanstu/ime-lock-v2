# IME Lock v2

`img-lock-v2` 是对 `D:\projects\ime-lock` 的一次独立重构版本，工程组织参考 `D:\projects\codexprov4`：使用 Go + Wails 提供桌面 UI，并把单实例、自启动、托盘、配置与核心监听拆成独立模块。

> 目录名按本次重构要求使用 `img-lock-v2`；应用显示名称仍为 **IME Lock v2**，因为实际功能是输入法（IME）状态守护。

## 功能

- 中文输入法落到英文子模式时自动恢复中文输入状态。
- `Ctrl + Shift + F9` 全局切换自动修复。
- Wails 状态面板：监听状态、累计恢复次数、最近恢复时间。
- 开机启动。
- 静默启动：仅开机自启时隐藏窗口；手动双击始终打开面板。
- 系统托盘驻留，关闭主窗口后继续运行。
- 再次双击 exe 不创建第二实例，而是唤起已经运行的主窗口。
- 可选日志采集，最多保留最近 600 条运行日志。

## 与原版 `ime-lock` 的主要区别

原版把 Win32 窗口、托盘、IAccessible 事件监听、配置、自启动、日志等全部放在一个约 1300 行的 `src/main.rs` 中。

v2 拆分为：

```text
img-lock-v2/
├── app.go                         # Wails App / 对前端暴露的 API
├── config.go                      # ~/.config/img-lock-v2/config.json
├── autostart_windows.go           # Windows Run 注册表
├── single_instance_windows.go     # 单实例 + 已有窗口唤醒
├── watcher_windows.go             # IME 核心检测与修复
├── tray_manager.go                # 系统托盘
├── frontend/src/                  # Wails UI
├── *_test.go                      # 基础回归测试
└── wails.json
```

### IME 检测策略

v2 保留原版已经验证有效的核心检测链路，只把它从 Rust/Win32 单文件中拆成独立的 Go 模块：

1. 在独立 Windows 消息线程中注册 `SetWinEventHook(EVENT_OBJECT_NAMECHANGE)`。
2. 通过 `AccessibleObjectFromEvent` 读取事件对象的 `IAccessible.accName`。
3. 仅处理包含“任务栏输入指示”的事件。
4. 明确检测到“英语模式”时，向当前前台窗口的默认 IME 窗口发送 `WM_IME_CONTROL / IMC_SETOPENSTATUS`，恢复中文模式。
5. 收到“中文模式”事件后清除 pending 状态，避免重复修复。

因此 Wails 只负责界面、托盘和配置，不参与 IME 状态判断；核心行为与原版 `ime-lock` 保持一致。

## 配置

配置文件：

```text
%USERPROFILE%\.config\img-lock-v2\config.json
```

当前持久化字段：

```json
{
  "auto_start": false,
  "silent_start": false
}
```

自动修复和日志采集是运行时状态，默认自动修复开启、日志采集关闭。

开机启动注册表项：

```text
HKCU\Software\Microsoft\Windows\CurrentVersion\Run
IMG-Lock-V2 = "<exe path>" --autostart
```

## 本地构建

需要：

- Go 1.25+
- Wails CLI v2.15+
- Windows 10/11
- WebView2 Runtime（Windows 11 通常已内置）

安装 Wails CLI：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
```

检查与构建：

```powershell
go fmt ./...
go mod tidy
go test ./...
wails build
```

构建产物：

```text
build\bin\img-lock-v2.exe
```

## 手工验证建议

1. 双击 `img-lock-v2.exe`，确认主面板显示且托盘图标存在。
2. 再双击一次，确认没有第二个进程，并且原窗口被唤起。
3. 在中文微软拼音下切到英文子模式，确认很快自动恢复中文。
4. 按 `Ctrl + Shift + F9`，确认自动修复开关状态变化。
5. 关闭主窗口，确认进程仍在托盘运行；点击托盘图标可重新打开。
6. 开启“开机启动 + 静默启动”，注销/重启后确认只启动托盘，不主动弹窗。
7. 已经静默启动后手动双击 exe，确认已有窗口会被主动唤起。
8. 开启日志采集，重复输入法切换，确认日志与恢复次数更新。

## CI

`.github/workflows/ci.yml` 在 Windows runner 上执行：

- `go test ./...`
- 安装 Wails CLI
- `wails build`
- 上传 `img-lock-v2.exe` 构建产物
