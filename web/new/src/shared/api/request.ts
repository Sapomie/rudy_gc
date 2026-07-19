import type { RouteLocationNormalizedLoaded } from "vue-router";
import { queryPart } from "@/shared/utils/format";

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function parseResponse(response: Response): Promise<unknown> {
  const raw = await response.text();
  let data: unknown = {};
  if (raw) {
    try {
      data = JSON.parse(raw);
    } catch {
      data = { error: raw };
    }
  }
  if (!response.ok) {
    const message =
      typeof data === "object" && data && "error" in data
        ? String((data as { error?: unknown }).error || response.statusText)
        : response.statusText;
    throw new ApiError(message, response.status);
  }
  return data;
}

export function apiPathForRoute(route: RouteLocationNormalizedLoaded): string {
  const configured = typeof route.meta.apiPath === "string" ? route.meta.apiPath : route.path;
  const explicit = configured.replace(/:([a-zA-Z0-9_]+)/g, (_, key: string) =>
    encodeURIComponent(String(route.params[key] || ""))
  );
  return `/api${explicit}${queryPart(route.fullPath)}`;
}

export async function getJSON<T = unknown>(path: string): Promise<T> {
  const response = await fetch(path, {
    headers: { Accept: "application/json" },
    credentials: "same-origin"
  });
  return (await parseResponse(response)) as T;
}

export async function sendJSON<T = unknown>(
  path: string,
  method: "POST" | "PUT" | "DELETE",
  payload?: unknown
): Promise<T> {
  const response = await fetch(path, {
    method,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json"
    },
    credentials: "same-origin",
    body: payload === undefined ? undefined : JSON.stringify(payload)
  });
  return (await parseResponse(response)) as T;
}
