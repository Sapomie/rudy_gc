import type { ApiEnvelope, ApiError } from "@/types/api";

const API_BASE = "/api/gc/v2";

export interface RequestOptions {
  params?: Record<string, unknown>;
  body?: unknown;
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
}

export async function apiGet<T>(path: string, params?: Record<string, unknown>): Promise<T> {
  return apiRequest<T>(path, { method: "GET", params });
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const method = options.method || "GET";
  const url = new URL(`${API_BASE}${path}`, window.location.origin);
  appendParams(url, options.params || {});

  const response = await fetch(url, {
    method,
    headers: options.body === undefined ? undefined : { "Content-Type": "application/json" },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  let envelope: ApiEnvelope<T>;
  try {
    envelope = (await response.json()) as ApiEnvelope<T>;
  } catch (error) {
    throw makeApiError("API 返回了无法解析的响应", response.status, error);
  }

  if (!response.ok || envelope.error) {
    throw makeApiError(envelope.error?.message || `请求失败：${response.status}`, response.status, envelope.error);
  }

  if (envelope.data === null) {
    throw makeApiError("API 响应缺少 data", response.status, envelope);
  }

  return envelope.data;
}

function appendParams(url: URL, params: Record<string, unknown>) {
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") {
      return;
    }
    url.searchParams.set(key, String(value));
  });
}

function makeApiError(message: string, status?: number, detail?: unknown): ApiError {
  const error = new Error(message) as ApiError;
  error.status = status;
  error.detail = detail;
  return error;
}
