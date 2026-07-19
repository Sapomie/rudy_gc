import type { ScActionResult, ScPageModel, ScRawRecord } from "./types";
import { mapActionResult, mapRecordPage, toRecord } from "@/shared/api/mappers";
export const mapScRecord = (value: unknown): ScRawRecord => toRecord(value);
export const mapScPage = (value: unknown): ScPageModel => mapRecordPage(value);
export const mapScAction = (value: unknown): ScActionResult => mapActionResult(value);
