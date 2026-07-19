import type { FC2ActionResult, FC2PageModel, FC2RawRecord } from "./types";
import { mapActionResult, mapRecordPage, toRecord } from "@/shared/api/mappers";
export const mapFC2Record = (value: unknown): FC2RawRecord => toRecord(value);
export const mapFC2Page = (value: unknown): FC2PageModel => mapRecordPage(value);
export const mapFC2Action = (value: unknown): FC2ActionResult => mapActionResult(value);
