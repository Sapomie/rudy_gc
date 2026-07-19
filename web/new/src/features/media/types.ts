export type MediaRecord = Record<string, unknown>;
export type MediaRawRecord = MediaRecord;
export interface MediaActionResult {
  success?: boolean;
  message?: string;
  error?: string;
}
export interface MediaPageModel {
  items: MediaRawRecord[];
  total?: number;
}
