export function estimateRemainingSeconds(elapsedSeconds: number, completed: number, total: number): number | null {
  if (![elapsedSeconds, completed, total].every(Number.isFinite) || total <= 0) {
    return null;
  }
  if (completed >= total) {
    return 0;
  }
  if (elapsedSeconds <= 0 || completed <= 0) {
    return null;
  }
  return Math.max(0, Math.round(elapsedSeconds * (total - completed) / completed));
}

export function formatTaskEta(seconds: number | null): string {
  if (seconds === null || !Number.isFinite(seconds) || seconds < 0) {
    return "计算中";
  }
  const value = Math.floor(seconds);
  const hours = Math.floor(value / 3600);
  const minutes = Math.floor((value % 3600) / 60);
  const remainingSeconds = value % 60;
  return [hours, minutes, remainingSeconds]
    .map((part) => String(part).padStart(2, "0"))
    .join(":");
}
