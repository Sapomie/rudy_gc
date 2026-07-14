import type { AnyRecord } from "@/shared/types/gc";
import { compactInternalHref } from "@/shared/utils/query";

export function asRecord(value: unknown): AnyRecord {
  return value && typeof value === "object" && !Array.isArray(value) ? value as AnyRecord : {};
}

export function asArray<T = unknown>(value: unknown): T[] {
  return Array.isArray(value) ? value as T[] : [];
}

export function pick(row: unknown, ...keys: string[]): unknown {
  const record = asRecord(row);
  for (const key of keys) {
    if (record[key] !== undefined && record[key] !== null) {
      return record[key];
    }
  }
  return undefined;
}

export function text(row: unknown, ...keys: string[]): string {
  const value = pick(row, ...keys);
  if (value === undefined || value === null) {
    return "";
  }
  return String(value);
}

export function num(row: unknown, ...keys: string[]): number {
  const value = Number(pick(row, ...keys) ?? 0);
  return Number.isFinite(value) ? value : 0;
}

export function bool(row: unknown, ...keys: string[]): boolean {
  return Boolean(pick(row, ...keys));
}

export function compact<T>(items: T[]): T[] {
  return items.filter((item) => item !== undefined && item !== null && String(item).trim() !== "");
}

export function formatValue(value: unknown): string {
  if (value === undefined || value === null || value === "") {
    return "-";
  }
  if (typeof value === "number") {
    return Number.isInteger(value) ? String(value) : value.toFixed(2);
  }
  if (typeof value === "boolean") {
    return value ? "是" : "否";
  }
  if (Array.isArray(value)) {
    return value.map((item) => formatValue(item)).join(", ");
  }
  if (typeof value === "object") {
    return JSON.stringify(value);
  }
  return String(value);
}

export function routeTitle(path: string): string {
  const clean = path.replace(/^\/+/, "").replaceAll("/", " / ");
  return clean || "Dashboard";
}

export function queryPart(fullPath: string): string {
  const idx = fullPath.indexOf("?");
  return idx >= 0 ? fullPath.slice(idx) : "";
}

export function normalizeHref(href: unknown, fallback = "#"): string {
  const value = String(href ?? "").trim();
  return compactInternalHref(value, fallback);
}
