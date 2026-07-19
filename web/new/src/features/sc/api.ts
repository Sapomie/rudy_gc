export { ApiError, apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
export const scApiPath = (path: string): string => `/api${path.startsWith("/") ? path : `/${path}`}`;
