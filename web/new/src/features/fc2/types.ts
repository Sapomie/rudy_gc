export type FC2Record = Record<string, unknown>;
export type FC2RawRecord = FC2Record;
export interface FC2ActionResult {
  success?: boolean;
  message?: string;
  error?: string;
}
export interface FC2PageModel {
  items: FC2RawRecord[];
  total?: number;
}
