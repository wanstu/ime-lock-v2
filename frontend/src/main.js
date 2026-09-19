const state = { current: null, loading: false };

const el = {
  message: document.getElementById("message"),
  runningPill: document.getElementById("running-pill"),
  autoFixStatus: document.getElementById("auto-fix-status"),
  fixCount: document.getElementById("fix-count"),
  lastFix: document.getElementById("last-fix"),
  autoFix: document.getElementById("auto-fix"),
  autoStart: document.getElementById("auto-start"),
  silentStart: document.getElementById("silent-start"),
  theme: document.getElementById("theme"),
  captureLogs: document.getElementById("capture-logs"),
  shortcut: document.getElementById("shortcut"),
  configPath: document.getElementById("config-path"),
  logOutput: document.getElementById("log-output"),
  clearLogs: document.getElementById("clear-logs"),
};

function api() {
  return window.go?.main?.App || null;
}

function errorText(error) {
  if (!error) return "发生未知错误";
  if (typeof error === "string") return error;
  if (error.message) return error.message;
  return String(error);
}

function showError(error) {
  el.message.textContent = errorText(error);
  el.message.classList.remove("hidden");
}

function clearError() {
  el.message.textContent = "";
  el.message.classList.add("hidden");
}

function normalizeTheme(mode) {
  return ["light", "dark", "system"].includes(mode) ? mode : "light";
}

function applyTheme(mode) {
  const nextMode = normalizeTheme(mode);
  const theme = window.desktopKitTheme;
  if (theme?.apply && theme?.getMode) {
    if (theme.getMode() !== nextMode) {
      theme.apply(nextMode);
    }
    return;
  }

  const resolved = nextMode === "system"
    ? (window.matchMedia?.("(prefers-color-scheme: dark)").matches ? "dark" : "light")
    : nextMode;
  document.documentElement.setAttribute("data-dk-theme", resolved);
  document.documentElement.setAttribute("data-dk-theme-mode", nextMode);
}

function render(next) {
  if (!next) return;
  state.current = next;

  el.runningPill.textContent = next.running ? "● 监听中" : "○ 未运行";
  el.runningPill.className = `status-pill dk-status-pill ${next.running ? "running is-success" : "stopped is-warning"}`;
  el.autoFixStatus.textContent = next.auto_fix ? "已开启" : "已关闭";
  el.fixCount.textContent = `${next.fix_count || 0} 次`;
  el.lastFix.textContent = next.last_fix_at || "暂无";

  el.autoFix.checked = Boolean(next.auto_fix);
  el.autoStart.checked = Boolean(next.auto_start);
  el.silentStart.checked = Boolean(next.silent_start);
  el.captureLogs.checked = Boolean(next.capture_logs);

  const theme = normalizeTheme(next.theme);
  el.theme.value = theme;
  applyTheme(theme);

  el.shortcut.textContent = next.shortcut || "Ctrl + Shift + F9";
  el.configPath.textContent = `配置路径：${next.config_path || "—"}`;
  el.configPath.title = next.config_path || "";

  if (next.last_error) {
    showError(next.last_error);
  } else {
    clearError();
  }
}

async function refresh() {
  const app = api();
  if (!app || state.loading) return;
  state.loading = true;
  try {
    const next = await app.GetState();
    render(next);
    if (next.capture_logs) {
      const logs = await app.GetLogs();
      el.logOutput.textContent = logs || "正在采集，暂时没有日志。";
      el.logOutput.scrollTop = el.logOutput.scrollHeight;
    } else {
      el.logOutput.textContent = "日志采集未开启。";
    }
  } catch (error) {
    showError(error);
  } finally {
    state.loading = false;
  }
}

function bindToggle(node, methodName) {
  node.addEventListener("change", async () => {
    const app = api();
    if (!app) return;
    const value = node.checked;
    node.disabled = true;
    try {
      const next = await app[methodName](value);
      render(next);
      clearError();
    } catch (error) {
      node.checked = !value;
      showError(error);
    } finally {
      node.disabled = false;
      await refresh();
    }
  });
}

bindToggle(el.autoFix, "SetAutoFix");
bindToggle(el.autoStart, "SetAutoStart");
bindToggle(el.silentStart, "SetSilentStart");
bindToggle(el.captureLogs, "SetCaptureLogs");

el.theme.addEventListener("change", async () => {
  const app = api();
  if (!app) return;

  const previous = normalizeTheme(state.current?.theme);
  const nextMode = normalizeTheme(el.theme.value);
  el.theme.disabled = true;
  applyTheme(nextMode);

  try {
    const next = await app.SetTheme(nextMode);
    render(next);
    clearError();
  } catch (error) {
    el.theme.value = previous;
    applyTheme(previous);
    showError(error);
  } finally {
    el.theme.disabled = false;
    await refresh();
  }
});

el.clearLogs.addEventListener("click", async () => {
  const app = api();
  if (!app) return;
  try {
    await app.ClearLogs();
    el.logOutput.textContent = state.current?.capture_logs ? "正在采集，暂时没有日志。" : "日志采集未开启。";
    clearError();
  } catch (error) {
    showError(error);
  }
});

async function bootstrap() {
  for (let i = 0; i < 50; i++) {
    if (api()) {
      await refresh();
      window.setInterval(refresh, 700);
      return;
    }
    await new Promise((resolve) => window.setTimeout(resolve, 100));
  }
  showError("Wails 后端未就绪，请重新启动应用。");
}

bootstrap();
