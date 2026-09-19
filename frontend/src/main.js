const state = {
  current: null,
  loading: false,
  themePacks: [],
  themeCatalog: null,
  themeCatalogLoading: false,
};

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
  themePack: document.getElementById("theme-pack"),
  themePackDescription: document.getElementById("theme-pack-description"),
  refreshThemePacks: document.getElementById("refresh-theme-packs"),
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

async function applyThemePack(pack) {
  const nextPack = String(pack || "aurora");
  const theme = window.desktopKitTheme;
  if (theme?.applyPack) {
    return theme.applyPack(nextPack);
  }
  if (theme?.setPack) {
    theme.setPack(nextPack);
  }
  return nextPack;
}

function effectiveThemePacks(selected) {
  const packs = Array.isArray(state.themePacks) ? state.themePacks.slice() : [];
  if (selected && !packs.some((pack) => pack.name === selected)) {
    packs.unshift({
      name: selected,
      display_name: selected,
      description: "当前已保存主题；远程主题目录中暂未找到该条目。",
    });
  }
  return packs;
}

function renderThemePacks(next) {
  const selected = next.theme_pack || "aurora";
  const packs = effectiveThemePacks(selected);
  el.themePack.replaceChildren(...packs.map((pack) => {
    const option = document.createElement("option");
    option.value = pack.name;
    option.textContent = pack.display_name || pack.name;
    return option;
  }));
  el.themePack.value = selected;

  const selectedPack = packs.find((pack) => pack.name === selected);
  if (state.themeCatalogLoading) {
    el.themePackDescription.textContent = "正在刷新 Wails Desktop Kit Theme 目录…";
  } else if (state.themeCatalog?.source === "builtin" && state.themeCatalog?.last_error) {
    el.themePackDescription.textContent = "远程主题不可用，当前使用 Kit 内置的 4 个基础配色主题。";
  } else if (state.themeCatalog?.source === "cache" && state.themeCatalog?.last_error) {
    el.themePackDescription.textContent = "远程主题刷新失败，当前使用本地主题缓存。";
  } else if (selectedPack) {
    el.themePackDescription.textContent = selectedPack.description || "由 Wails Desktop Kit Theme Runtime 提供。";
  } else {
    el.themePackDescription.textContent = "主题目录暂不可用；当前继续使用 Kit 基础主题。";
  }

  applyThemePack(selected).catch(() => {
    el.themePackDescription.textContent = "当前配色加载失败，已继续使用 Kit 基础主题。";
  });
}

async function loadThemeCatalog(refresh = false, reportError = false) {
  const theme = window.desktopKitTheme;
  if (!theme?.loadCatalog || state.themeCatalogLoading) return;

  state.themeCatalogLoading = true;
  el.refreshThemePacks.disabled = true;
  if (state.current) renderThemePacks(state.current);

  try {
    const catalog = refresh && theme.refreshCatalog
      ? await theme.refreshCatalog()
      : await theme.loadCatalog();
    state.themeCatalog = catalog;
    state.themePacks = Array.isArray(catalog?.packs) ? catalog.packs : [];
    if (state.current) renderThemePacks(state.current);
  } catch (error) {
    if (state.current) renderThemePacks(state.current);
    el.themePackDescription.textContent = state.themePacks.length
      ? "远程主题目录刷新失败，正在使用本地缓存。"
      : "主题目录暂不可用；当前继续使用 Kit 基础主题。";
    if (reportError) showError(error);
  } finally {
    state.themeCatalogLoading = false;
    el.refreshThemePacks.disabled = false;
    if (state.current) renderThemePacks(state.current);
  }
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
  renderThemePacks(next);

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

el.themePack.addEventListener("change", async () => {
  const app = api();
  if (!app) return;

  const previous = state.current?.theme_pack || "aurora";
  const nextPack = el.themePack.value;
  el.themePack.disabled = true;

  try {
    await applyThemePack(nextPack);
    const next = await app.SetThemePack(nextPack);
    render(next);
    clearError();
  } catch (error) {
    el.themePack.value = previous;
    await applyThemePack(previous).catch(() => {});
    showError(error);
  } finally {
    el.themePack.disabled = false;
    await refresh();
  }
});

el.refreshThemePacks.addEventListener("click", async () => {
  clearError();
  await loadThemeCatalog(true, true);
});

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
      await loadThemeCatalog(true, false);
      window.setInterval(refresh, 700);
      return;
    }
    await new Promise((resolve) => window.setTimeout(resolve, 100));
  }
  showError("Wails 后端未就绪，请重新启动应用。");
}

bootstrap();
