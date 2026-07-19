export type CrawlerRecord = Record<string, unknown>;
export type CrawlerRawRecord = CrawlerRecord;
export interface CrawlerActionResult {
  success?: boolean;
  message?: string;
  error?: string;
}
export interface CrawlerPageModel {
  items: CrawlerRawRecord[];
  total?: number;
}
