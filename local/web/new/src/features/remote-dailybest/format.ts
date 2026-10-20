const dateTimeFormatter = new Intl.DateTimeFormat("zh-CN", {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hour12: false
});

export function formatDateTime(unixSeconds: number): string {
  if (!unixSeconds) return "—";
  return dateTimeFormatter.format(new Date(unixSeconds * 1000));
}

export function formatCategory(category: number): string {
  return category === 1 ? "月榜" : String(category);
}

export function formatRunStatus(status: string): string {
  return (
    {
      queued: "等待执行",
      running: "执行中",
      succeeded: "已发布",
      failed: "失败",
      no_change: "内容未变化"
    }[status] || status
  );
}

export function formatFetchLogStatus(status: string): string {
  return { requesting: "请求中", succeeded: "成功", failed: "失败" }[status] || status;
}

export function formatRankDaySyncStatus(status: string): string {
  return (
    {
      missing: "本地缺失",
      incomplete: "数量不完整",
      version_mismatch: "版本不一致",
      pending: "待处理",
      failed: "处理失败",
      completed: "已完成"
    }[status] ||
    status ||
    "未知"
  );
}

export function rankDaySyncStatusTone(status: string): "good" | "warn" | "danger" | "info" | "neutral" {
  if (status === "completed") return "good";
  if (status === "failed") return "danger";
  if (status === "pending") return "info";
  if (status === "missing" || status === "incomplete" || status === "version_mismatch") return "warn";
  return "neutral";
}

export function formatTriggerType(triggerType: string): string {
  return { scheduled: "定时任务", recovery: "失败恢复", manual: "手动触发" }[triggerType] || triggerType;
}

export function formatRuntimeState(status: string): string {
  return (
    {
      starting: "启动中",
      ready: "就绪",
      fetching: "请求中",
      disconnected: "连接断开",
      recovering: "恢复中",
      fatal: "严重故障",
      available: "可访问",
      unavailable: "不可访问",
      unknown: "尚未探测",
      interrupted: "已中断"
    }[status] ||
    status ||
    "—"
  );
}

export function formatProbeStatus(status: string): string {
  return (
    {
      running: "探测中",
      succeeded: "成功",
      failed: "失败",
      interrupted: "已中断"
    }[status] || status
  );
}

export function statusTone(status: string): string {
  if (status === "succeeded" || status === "ready" || status === "available") return "good";
  if (status === "failed" || status === "unavailable" || status === "fatal" || status === "interrupted") {
    return "danger";
  }
  if (status === "running" || status === "requesting" || status === "starting" || status === "fetching") return "info";
  if (status === "queued" || status === "disconnected" || status === "recovering" || status === "unknown")
    return "warn";
  return "neutral";
}

export function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
