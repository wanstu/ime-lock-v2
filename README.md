# IME Lock v2

`ime-lock-v2` 是 Windows 输入法状态守护工具：当中文输入法意外落到英文子模式时，自动恢复中文输入状态。

桌面基础设施已迁移到 **Wails Desktop Kit v0.2.2**。IME 检测与修复仍由本项目负责；Wails 生命周期、托盘、单实例、开机启动、共享 UI 资源、主题机制和图标规范化由 Kit 提供。

## 功能

- 中文输入法落到英文子模式时自动恢复中文输入状态。
- `Ctrl + Shift + F9` 全局切换自动修复。
- Wails 状态面板：监听状态、累计恢复次数、最近恢复时间。
- 开机启动。
- 静默启动：仅开机自启时在托盘就绪后隐藏窗口；手动启动始终显示面板。
- 主题：浅色、深色、跟随系统；默认保持浅色，选择会持久化。
- 系统托盘驻留，关闭主窗口后继续运行。
- 单实例：
  - 手动重复启动会唤起已运行窗口。
  - `--autostart` / 旧版 `--minimized` 重复启动会静默退出。
- 可选日志采集，最多保留最近 600 条运行日志。

## Desktop Kit 迁移边界

Kit 接管：

- Wails `Run` 生命周期和窗口隐藏/恢复策略。
- 声明式托盘菜单和托盘错误恢复。
- Wails 单实例锁与第二实例回调。
- Windows 登录自启动管理。
- `/desktopkit/` 共享 UI CSS。
- `light / dark / system` 主题解析、system 监听与语义 token。
- 应用图标规范化 CLI。
- Release 产物命名和 SHA256 约定。

IME Lock 保留：

- Windows `SetWinEventHook` / `IAccessible` 检测。
- `WM_IME_CONTROL / IMC_SETOPENSTATUS` 修复逻辑。
- 全局快捷键。
- 配置、日志和业务状态。
- 主题设置 UI、默认值和持久化；Kit 只负责通用主题机制和 token。
- Windows-only CI 矩阵。核心 IME 监听并不支持 Linux/macOS，因此不会为了统一 CI 而发布无实际功能的跨平台产物。

当前主要结构：

```text
ime-lock-v2/
├── main.go                         # Desktop Kit Runtime 入口
├── desktop_runtime.go              # Kit 托盘、单实例策略
├── app.go                          # 前端 API / 业务状态
├── config.go                       # ~/.config/ime-lock-v2/config.json
├── autostart_legacy_windows.go     # 仅负责清理历史 IMG-Lock-V2 项
├── watcher_windows.go              # IME 核心检测与修复
├── frontend/src/                  # Wails UI + Kit CSS
├── scripts/generate-app-icon.ps1  # 调用 Desktop Kit icon CLI
└── wails.json
```

旧的自定义 `TrayManager`、文件锁单实例 IPC 和完整 Windows Run 注册表实现已经移除。

## IME 检测策略

核心检测链路保持不变：

1. 在独立 Windows 消息线程中注册 `SetWinEventHook(EVENT_OBJECT_NAMECHANGE)`。
2. 通过 `AccessibleObjectFromEvent` 读取事件对象的 `IAccessible.accName`。
3. 仅处理包含“任务栏输入指示”的事件。
4. 明确检测到“英语模式”时，向当前前台窗口的默认 IME 窗口发送 `WM_IME_CONTROL / IMC_SETOPENSTATUS`，恢复中文模式。
5. 收到“中文模式”事件后清除 pending 状态，避免重复修复。

Desktop Kit 不参与 IME 状态判断。

## 配置与开机启动

配置文件：

```text
%USERPROFILE%\.config\ime-lock-v2\config.json
```

当前持久化字段：

```json
{
  "auto_start": false,
  "silent_start": false,
  "theme": "light"
}
```

开机启动仍沿用原有 Windows Run value 名称，避免升级改变用户已有设置：

```text
HKCU\Software\Microsoft\Windows\CurrentVersion\Run
IME-Lock-V2 = "<exe path>" --autostart
```

历史错误名称 `IMG-Lock-V2` 会在更新开机启动设置时清理。

`theme` 可取 `light`、`dark`、`system`。旧配置没有该字段时按 `light` 处理，避免升级后外观突然变化；选择 `system` 后由 Kit 的 `theme.js` 监听 Windows 系统主题变化。

## 本地构建

需要：

- Go 1.26+
- Wails CLI v2.15+
- Windows 10/11
- WebView2 Runtime

安装 Wails CLI：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
```

检查与构建：

```powershell
go fmt ./...
go mod tidy
go test ./...
go vet ./...
wails build -clean
```

构建时 `scripts/generate-app-icon.ps1` 会调用当前 Go Module 固定的 Desktop Kit icon CLI 生成 `build/appicon.png`。

最终产物：

```text
build\bin\ime-lock-v2.exe
```

## 手工验证建议

1. 双击 `ime-lock-v2.exe`，确认主面板显示且托盘图标存在。
2. 再双击一次，确认没有第二个实例，并且原窗口被唤起。
3. 用 `ime-lock-v2.exe --autostart` 模拟重复登录启动，确认已有实例不被主动弹窗。
4. 兼容检查 `--minimized`，行为应与 `--autostart` 一致。
5. 在中文微软拼音下切到英文子模式，确认自动恢复中文。
6. 按 `Ctrl + Shift + F9`，确认自动修复开关状态变化。
7. 关闭主窗口，确认进程仍驻留托盘；托盘可恢复窗口。
8. 开启“开机启动 + 静默启动”，注销/重启后确认托盘就绪后自动隐藏。
9. 分别切换浅色、深色、跟随系统，确认公共组件和 IME 自定义区域均正确换色；选择跟随系统后切换 Windows 主题，页面应实时更新。
10. 重启应用，确认主题选择仍然保留；旧配置升级后的首次启动仍保持浅色。
11. 分别从前端和托盘切换自动修复、开机启动、静默启动、日志采集，确认状态一致。
12. 开启日志采集，重复输入法切换，确认日志与恢复次数更新。

## CI / Release

IME Lock 保持 Windows-only workflow，但发布约定与 Desktop Kit 对齐：

- `go test ./...`
- `go vet ./...`
- `wails build -clean`
- 普通 CI 产物：`ime-lock-v2-windows-amd64.exe`
- tag `vX.Y.Z`：`ime-lock-v2-vX.Y.Z-windows-amd64.exe`
- 同时发布对应 `.sha256` 文件。

Linux/macOS 不进入 IME Lock Release，因为当前核心监听明确只支持 Windows。
