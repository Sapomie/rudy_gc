export { ApiError, apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
export const movieApiPath = (path: string): string => `/api${path.startsWith("/") ? path : `/${path}`}`;
