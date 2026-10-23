import { getJSON, sendJSON } from "@/shared/api/request";

export { ApiError, apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
export const movieApiPath = (path: string): string => `/api${path.startsWith("/") ? path : `/${path}`}`;

export interface CafoProfile {
  id: number;
  personId: number;
  name: string;
  alias: string;
  chinese: string;
  birthDay: number;
  height: number;
  cup: string;
  bwh: string;
  avatar: string;
  biography: string;
  createdOn: number;
  updatedOn: number;
}

export interface CafoWriteInput {
  name: string;
  alias: string;
  chinese: string;
  birthDay: number;
  height: number;
  cup: string;
  bwh: string;
  avatar: string;
  biography: string;
}

export interface CafoUpdateResponse {
  ok: boolean;
  message: string;
  cafo: CafoProfile;
  person: Record<string, unknown>;
}

export async function listPersonCafos(personID: number): Promise<CafoProfile[]> {
  const response = await getJSON<{ cafos?: CafoProfile[] }>(personCafoPath(personID));
  return Array.isArray(response.cafos) ? response.cafos : [];
}

export function updatePersonCafo(personID: number, cafoID: number, input: CafoWriteInput): Promise<CafoUpdateResponse> {
  return sendJSON<CafoUpdateResponse>(personCafoPath(personID, cafoID), "PUT", input);
}

function personCafoPath(personID: number, cafoID?: number): string {
  const base = `/api/persons/${encodeURIComponent(personID)}/cafos`;
  return cafoID === undefined ? base : `${base}/${encodeURIComponent(cafoID)}`;
}
