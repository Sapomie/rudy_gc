import type { CrawlerActionResult, CrawlerPageModel, CrawlerRawRecord } from "./types";
import { mapActionResult, mapRecordPage, toRecord } from "@/shared/api/mappers";
export const mapCrawlerRecord = (value: unknown): CrawlerRawRecord => toRecord(value);
export const mapCrawlerPage = (value: unknown): CrawlerPageModel => mapRecordPage(value);
export const mapCrawlerAction = (value: unknown): CrawlerActionResult => mapActionResult(value);
