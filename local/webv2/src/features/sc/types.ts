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

export type ScPickV2FilterValue = string | number | null;

export interface ScPickV2Filter {
  [key: string]: ScPickV2FilterValue;
}

export interface ScPickV2Group {
  id: string;
  weight: number;
  expanded: boolean;
  filter: ScPickV2Filter;
}

export interface ScPickV2JAVOptions {
  castParticipatedScBlockDays: number;
  castSelectedScBlockDays: number;
  rank20Min: number;
  rank100Min: number;
  rank500Min: number;
  castParticipatedScPenaltyAlpha: number;
  castSelectedScPenaltyAlpha: number;
  castOwnedScRatioPenaltyAlpha: number;
  movieHasSelectedScPenaltyAlpha: number;
  randomSeed?: number;
}

export interface ScPickV2FC2Options {
  movieParticipatedScBlockDays: number;
  movieSelectedScBlockDays: number;
  movieSelectedScPenaltyAlpha: number;
  movieHasSelectedScPenaltyAlpha: number;
  randomSeed?: number;
}

export interface ScPickV2Request {
  source: "wmedia";
  jav: {
    pickN: number;
    options: ScPickV2JAVOptions;
    reqs: Array<{ weight: number; req: ScPickV2Filter }>;
  };
  fc2: {
    pickN: number;
    options: ScPickV2FC2Options;
    reqs: Array<{ weight: number; fc2Req: ScPickV2Filter }>;
  };
}

export interface ScPickV2Cast {
  person_id: number;
  name: string;
}

export interface ScPickV2JAVMovie {
  [key: string]: unknown;
  name: string;
  title: string;
  releasing_date: string;
  film_birth_date_w_media: string;
  video_url_w_media: string;
  owned_w_media: number;
  cast: ScPickV2Cast[];
  picked_order: number;
}

export interface ScPickV2Section<T> {
  picked: number;
  movies: T[];
  totalSizeBytes: number;
  totalSizeGB: string;
  pickInfo: Record<string, number | string>;
}

export interface ScPickV2Response {
  previewID: string;
  picked: number;
  totalSizeBytes: number;
  totalSizeGB: string;
  jav: ScPickV2Section<ScPickV2JAVMovie>;
  fc2: ScPickV2Section<ScRecord>;
}

export type ScCopyV2State = "running" | "stopping" | "completed" | "failed" | "stopped";

export interface ScCopyV2Failure {
  fileName: string;
  error: string;
}

export interface ScCopyV2Status {
  jobID: string;
  state: ScCopyV2State;
  total: number;
  done: number;
  success: number;
  failed: number;
  skipped: number;
  current: string;
  totalBytes: number;
  copiedBytes: number;
  startedAt: number;
  finishedAt: number;
  stopRequested: boolean;
  failures: ScCopyV2Failure[];
}
