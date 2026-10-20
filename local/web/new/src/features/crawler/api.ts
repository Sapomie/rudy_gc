export { ApiError, apiPathForRoute, getJSON, sendJSON } from "@/shared/api/request";
export const crawlerApiPath = (path: string): string => `/api${path.startsWith("/") ? path : `/${path}`}`;
