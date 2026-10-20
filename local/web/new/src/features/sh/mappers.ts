import type { SHActionResult, SHPageModel, SHRawRecord } from "./types";
import { mapActionResult, mapRecordPage, toRecord } from "@/shared/api/mappers";
export const mapSHRecord = (value: unknown): SHRawRecord => toRecord(value);
export const mapSHPage = (value: unknown): SHPageModel => mapRecordPage(value);
export const mapSHAction = (value: unknown): SHActionResult => mapActionResult(value);
