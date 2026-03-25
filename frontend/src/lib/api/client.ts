import { camelizeKeysDeep } from "@/src/lib/api/normalize";
import { SafeJson } from "../utils";

export type ClientError = {
  code: string;
  message: string;
  status: number;
};

export async function apiGet<T>(path: string): Promise<T> {
  return apiRequest<T>({ method: "GET", path });
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>({ method: "POST", path, body });
}

export async function apiPatch<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>({ method: "PATCH", path, body });
}

export async function apiPut<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>({ method: "PUT", path, body });
}

export async function apiDelete<T>(path: string): Promise<T> {
  return apiRequest<T>({ method: "DELETE", path });
}

async function apiRequest<T>(input: { method: string; path: string; body?: unknown }): Promise<T> {
  const resp = await fetch(`/api/gateway/${stripLeadingSlash(input.path)}`, {
    method: input.method,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json"
    },
    body: input.body === undefined ? undefined : JSON.stringify(input.body)
  });

  const text = await resp.text();
  const json = text ? SafeJson(text) : undefined;

  if (!resp.ok) {
    const code = typeof (json as any)?.code === "string" ? (json as any).code : "HTTP_ERROR";
    const message =
      typeof (json as any)?.message === "string" ? (json as any).message : `Request failed (${resp.status})`;
    const err: ClientError = { code, message, status: resp.status };
    throw err;
  }

  return camelizeKeysDeep<T>(json);
}

function stripLeadingSlash(path: string): string {
  return path.startsWith("/") ? path.slice(1) : path;
}
