import type { MediaActionResult, MediaPageModel, MediaRawRecord } from "./types";
import { mapActionResult, mapRecordPage, toRecord } from "@/shared/api/mappers";
export const mapMediaRecord = (value: unknown): MediaRawRecord => toRecord(value);
export const mapMediaPage = (value: unknown): MediaPageModel => mapRecordPage(value);
export const mapMediaAction = (value: unknown): MediaActionResult => mapActionResult(value);
