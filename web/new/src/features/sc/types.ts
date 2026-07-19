export type ScRecord = Record<string, unknown>;
export type ScRawRecord = ScRecord;
export interface ScActionResult {
  success?: boolean;
  message?: string;
  error?: string;
}
export interface ScPageModel {
  items: ScRawRecord[];
  total?: number;
}
