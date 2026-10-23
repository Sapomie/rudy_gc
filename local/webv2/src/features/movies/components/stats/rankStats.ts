import type { AnyRecord } from "@/shared/types/gc";
import { num } from "@/shared/utils/format";

export function formatRankStatCount(value: unknown): string {
  return new Intl.NumberFormat("zh-CN").format(Number(value || 0));
}

export function formatRankStatDecimal(value: unknown): string {
  return Number(value || 0).toFixed(2);
}

export function rankDateFromDayNumber(value: unknown, zeroIsEmpty = true): string {
  const dayNumber = Number(value);
  if (!Number.isFinite(dayNumber) || (zeroIsEmpty && dayNumber === 0)) return "-";
  const date = new Date(Date.UTC(2023, 0, 1) + (dayNumber - 1) * 86_400_000);
  return date.toISOString().slice(0, 10);
}

export function rankPeriodDateRange(item: AnyRecord): string {
  return `${rankDateFromDayNumber(num(item, "StartDayNumber"), false)} — ${rankDateFromDayNumber(num(item, "EndDayNumber"), false)}`;
}
