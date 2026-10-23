import { getJSON, sendFormData, sendJSON } from "@/shared/api/request";
import { asArray, asRecord, num, pick, text } from "@/shared/utils/format";
import type {
  ScCopyV2State,
  ScCopyV2Status,
  ScPickV2JAVMovie,
  ScPickV2Request,
  ScPickV2Response,
  ScPickV2Section,
  ScRecord
} from "@/features/sc/types";

export { ApiError, apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";

export function deleteScEvent(name: string): Promise<unknown> {
  return sendJSON(`/api/sc-events/${encodeURIComponent(name)}`, "DELETE");
}

export interface ScEventUpdateInput {
  comeGListId: number;
  movieCastPersonId: number;
  kindOptionId: number;
  duration: number;
  fgOptionId: number;
  vesselOptionId: number;
  remarks: string;
}

export function updateScEvent(name: string, input: ScEventUpdateInput, image?: File | null): Promise<unknown> {
  const path = `/api/sc-events/${encodeURIComponent(name)}`;
  if (!image) return sendJSON(path, "POST", input);
  const data = new FormData();
  data.append("data", JSON.stringify(input));
  data.append("image", image);
  return sendFormData(path, "POST", data);
}

export type ScOptionType = "kind" | "fg" | "vessel";

export interface ScOptionItem {
  id: number;
  optionType: ScOptionType;
  optionName: string;
  sortOrder: number;
  isEnabled: number;
  scEventTimes: number;
  lastScEventTime: number;
  createdOn: number;
  updatedOn: number;
}

export interface ScOptionGroups {
  kind: ScOptionItem[];
  fg: ScOptionItem[];
  vessel: ScOptionItem[];
}

export interface ScOptionWriteInput {
  optionName: string;
  sortOrder: number;
  isEnabled: number;
}

export async function getScOptions(includeDisabled = false): Promise<ScOptionGroups> {
  const path = includeDisabled ? "/api/sc-options?includeDisabled=1" : "/api/sc-options";
  const response = asRecord(await getJSON(path));
  return {
    kind: mapScOptionItems(pick(response, "kind"), "kind"),
    fg: mapScOptionItems(pick(response, "fg"), "fg"),
    vessel: mapScOptionItems(pick(response, "vessel"), "vessel")
  };
}

export function createScOption(optionType: ScOptionType, input: ScOptionWriteInput): Promise<unknown> {
  return sendJSON(scOptionPath(optionType), "POST", input);
}

export function updateScOption(optionType: ScOptionType, id: number, input: ScOptionWriteInput): Promise<unknown> {
  return sendJSON(scOptionPath(optionType, id), "PUT", input);
}

export function deleteScOption(optionType: ScOptionType, id: number): Promise<unknown> {
  return sendJSON(scOptionPath(optionType, id), "DELETE");
}

function mapScOptionItems(value: unknown, optionType: ScOptionItem["optionType"]): ScOptionItem[] {
  return asArray(value).map((value) => {
    const row = asRecord(value);
    return {
      id: num(row, "id", "ID"),
      optionType,
      optionName: text(row, "optionName", "OptionName"),
      sortOrder: num(row, "sortOrder", "SortOrder"),
      isEnabled: num(row, "isEnabled", "IsEnabled"),
      scEventTimes: num(row, "scEventTimes", "ScEventTimes"),
      lastScEventTime: num(row, "lastScEventTime", "LastScEventTime"),
      createdOn: num(row, "createdOn", "CreatedOn"),
      updatedOn: num(row, "updatedOn", "UpdatedOn")
    };
  });
}

function scOptionPath(optionType: ScOptionType, id?: number): string {
  const base = `/api/sc-options/${encodeURIComponent(optionType)}`;
  return id === undefined ? base : `${base}/${encodeURIComponent(id)}`;
}
export const scApiPath = (path: string): string => `/api${path.startsWith("/") ? path : `/${path}`}`;

export async function runScPickV2(request: ScPickV2Request): Promise<ScPickV2Response> {
  return mapScPickV2Response(await sendJSON("/api/triggers/sc/v2/pick", "POST", request));
}

export async function startScPickV2Copy(previewID: string): Promise<ScCopyV2Status> {
  return mapScCopyV2Status(await sendJSON("/api/triggers/sc/v2/copy", "POST", { preview_id: previewID }));
}

export async function getScPickV2CopyStatus(jobID: string): Promise<ScCopyV2Status> {
  return mapScCopyV2Status(await getJSON(`/api/triggers/sc/v2/copy/${encodeURIComponent(jobID)}`));
}

export async function stopScPickV2Copy(jobID: string): Promise<ScCopyV2Status> {
  return mapScCopyV2Status(await sendJSON(`/api/triggers/sc/v2/copy/${encodeURIComponent(jobID)}/stop`, "POST", {}));
}

function mapScPickV2Response(value: unknown): ScPickV2Response {
  const row = asRecord(value);
  return {
    previewID: text(row, "preview_id"),
    picked: num(row, "picked"),
    totalSizeBytes: num(row, "total_size_bytes"),
    totalSizeGB: text(row, "total_size_gb"),
    jav: mapScPickV2Section(asRecord(pick(row, "jav")), true) as ScPickV2Section<ScPickV2JAVMovie>,
    fc2: mapScPickV2Section(asRecord(pick(row, "fc2")), false)
  };
}

function mapScPickV2Section(row: ScRecord, jav: boolean): ScPickV2Section<ScRecord> {
  const movies = asArray<ScRecord>(pick(row, "movies")).map((movie, index) => ({
    ...movie,
    picked_order: num(movie, "picked_order") || index
  }));
  return {
    picked: num(row, "picked"),
    movies: jav ? movies.map((movie) => ({ ...movie })) : movies,
    totalSizeBytes: num(row, "total_size_bytes"),
    totalSizeGB: text(row, "total_size_gb"),
    pickInfo: asRecord(pick(row, "pick_info")) as Record<string, number | string>
  };
}

function mapScCopyV2Status(value: unknown): ScCopyV2Status {
  const row = asRecord(value);
  return {
    jobID: text(row, "job_id"),
    state: (text(row, "state") || "running") as ScCopyV2State,
    total: num(row, "total"),
    done: num(row, "done"),
    success: num(row, "success"),
    failed: num(row, "failed"),
    skipped: num(row, "skipped"),
    current: text(row, "current"),
    totalBytes: num(row, "total_bytes"),
    copiedBytes: num(row, "copied_bytes"),
    startedAt: num(row, "started_at"),
    finishedAt: num(row, "finished_at"),
    stopRequested: Boolean(pick(row, "stop_requested")),
    failures: asArray(pick(row, "failures")).map((item) => {
      const failure = asRecord(item);
      return { fileName: text(failure, "file_name"), error: text(failure, "error") };
    })
  };
}
