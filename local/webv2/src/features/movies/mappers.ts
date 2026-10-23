import type { MovieActionResult, MoviePageModel, MovieRawRecord } from "./types";
import { mapActionResult, mapRecordPage, toRecord } from "@/shared/api/mappers";
export const mapMovieRecord = (value: unknown): MovieRawRecord => toRecord(value);
export const mapMoviePage = (value: unknown): MoviePageModel => mapRecordPage(value);
export const mapMovieAction = (value: unknown): MovieActionResult => mapActionResult(value);
