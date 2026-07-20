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

export function rankDate(dayNumber: number): string {
  if (!dayNumber) return "—";
  const date = new Date(Date.UTC(2023, 0, dayNumber));
  return date.toISOString().slice(0, 10);
}

export function formatPeriodType(periodType: number): string {
  return ({ 1: "周", 2: "月", 3: "季度", 4: "年" } as Record<number, string>)[periodType] || String(periodType);
}

export function formatPeriodStatus(status: number): string {
  return ({ 1: "计算中", 2: "已就绪" } as Record<number, string>)[status] || String(status);
}

export function formatDecimal(value: number): string {
  return Number.isFinite(value) ? value.toFixed(4) : "—";
}

export function formatRankChange(value: number): string {
  if (value > 0) return `+${value}`;
  if (value < 0) return String(value);
  return "—";
}

export function rankChangeClass(value: number): string {
  if (value > 0) return "rank-change-rise";
  if (value < 0) return "rank-change-fall";
  return "";
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

export function formatTriggerType(triggerType: string): string {
  return { scheduled: "定时任务", manual: "手动触发" }[triggerType] || triggerType;
}

export function statusTone(status: string): string {
  if (status === "succeeded") return "good";
  if (status === "failed") return "danger";
  if (status === "running") return "info";
  if (status === "queued") return "warn";
  return "neutral";
}

export function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
