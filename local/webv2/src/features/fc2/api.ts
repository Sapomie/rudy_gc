export { ApiError, apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
export const fc2ApiPath = (path: string): string => `/api${path.startsWith("/") ? path : `/${path}`}`;
