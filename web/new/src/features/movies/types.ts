export type MovieRecord = Record<string, unknown>;
export type MovieRawRecord = MovieRecord;
export interface MovieActionResult {
  success?: boolean;
  message?: string;
  error?: string;
}
export interface MoviePageModel {
  items: MovieRawRecord[];
  total?: number;
}
