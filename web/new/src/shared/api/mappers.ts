export interface RecordPageModel<T extends Record<string, unknown> = Record<string, unknown>> {
  items: T[];
  total?: number;
}

export interface ActionResultModel {
  success?: boolean;
  message?: string;
  error?: string;
}

export function toRecord(value: unknown): Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}

export function pickField(record: Record<string, unknown>, ...keys: string[]): unknown {
  for (const key of keys) {
    if (record[key] !== undefined && record[key] !== null) {
      return record[key];
    }
  }
  return undefined;
}

export function stringField(record: Record<string, unknown>, ...keys: string[]): string {
  const value = pickField(record, ...keys);
  return value === undefined ? "" : String(value);
}

export function mapRecordPage(value: unknown): RecordPageModel {
  const record = toRecord(value);
  const rawItems = pickField(record, "items", "Items", "rows", "Rows", "data", "Data");
  const items = Array.isArray(rawItems) ? rawItems.map(toRecord) : [];
  const rawTotal = pickField(record, "total", "Total", "count", "Count");
  const parsedTotal = rawTotal === undefined ? undefined : Number(rawTotal);
  return { items, total: Number.isFinite(parsedTotal) ? parsedTotal : undefined };
}

export function mapActionResult(value: unknown): ActionResultModel {
  const record = toRecord(value);
  const success = pickField(record, "success", "Success", "ok", "OK");
  return {
    success: success === undefined ? undefined : Boolean(success),
    message: stringField(record, "message", "Message"),
    error: stringField(record, "error", "Error")
  };
}
