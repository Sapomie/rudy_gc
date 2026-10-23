export type ScDurationInput = number | "";

export function normalizeScDuration(value: ScDurationInput): number {
  if (value === "") {
    return 0;
  }
  const duration = Number.parseInt(String(value), 10);
  if (!Number.isFinite(duration)) {
    throw new Error("时长(分钟)必须是数字");
  }
  return duration;
}
