export type SHRecord = Record<string, unknown>;
export type SHRawRecord = SHRecord;
export interface SHActionResult {
  success?: boolean;
  message?: string;
  error?: string;
}
export interface SHPageModel {
  items: SHRawRecord[];
  total?: number;
}
