import { camelizeKeysDeep } from "@/src/lib/api/normalize";
import { SafeJson } from "@/src/lib/utils";
import type { TokenResponse } from "@/src/lib/api/backend-types";
import {
  getConsentModeClient,
  getStoredTokensForMode,
  saveLocalTokens,
  setSessionExpiresAtCookie,
} from "@/src/services/authStorage";

export type ClientError = {
  code: string;
  message: string;
  status: number;
};

function stripLeadingSlash(path: string): string {
  return path.startsWith("/") ? path.slice(1) : path;
}

async function parseError(resp: Response): Promise<ClientError> {
  const text = await resp.text();
  const json = text ? (SafeJson(text) as any) : undefined;

  const code = typeof json?.code === "string" && json.code ? json.code : "HTTP_ERROR";
  const message =
    typeof json?.message === "string" && json.message ? json.message : `Request failed (${resp.status})`;

  return { code, message, status: resp.status };
}

async function refreshOnce(storageMode: "cookie" | "localStorage") {
  if (storageMode === "localStorage") {
    const { refreshToken } = getStoredTokensForMode("localStorage");
    if (!refreshToken) throw new Error("no refresh token");

    const resp = await fetch("/api/auth/refresh", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "x-velune-refresh-token": refreshToken,
      },
    });

    if (!resp.ok) throw new Error("refresh failed");

    const tokens = (await resp.json()) as TokenResponse;
    if (!tokens.access_token || !tokens.refresh_token) throw new Error("invalid tokens");

    saveLocalTokens(tokens);
    setSessionExpiresAtCookie(tokens.expires_in);
    return tokens.access_token;
  }

  // Cookie mode: refresh endpoint uses httpOnly cookies.
  const resp = await fetch("/api/auth/refresh", { method: "POST" });
  if (!resp.ok) throw new Error("refresh failed");
  return null;
}

export async function apiRequest<T>(input: {
  method: string;
  path: string;
  body?: unknown;
  headers?: Record<string, string>;
}): Promise<T> {
  const storageMode = getConsentModeClient() ?? "cookie";
  const isLocal = storageMode === "localStorage";

  const { accessToken } = isLocal ? getStoredTokensForMode("localStorage") : { accessToken: null };

  const gatewayPath = `/api/gateway/${stripLeadingSlash(input.path)}`;

  const headers: Record<string, string> = {
    Accept: "application/json",
    "Content-Type": "application/json",
    ...(input.headers ?? {}),
  };

  if (isLocal && accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }

  const resp = await fetch(gatewayPath, {
    method: input.method,
    headers,
    body: input.body === undefined ? undefined : JSON.stringify(input.body),
  });

  if (resp.ok) {
    const text = await resp.text();
    const json = text ? SafeJson(text) : undefined;
    return camelizeKeysDeep<T>(json);
  }

  if (resp.status === 401) {
    try {
      await refreshOnce(storageMode);

      const tokensAfterRefresh = isLocal ? getStoredTokensForMode("localStorage") : { accessToken: null };
      const retryHeaders: Record<string, string> = { ...headers };
      if (isLocal && tokensAfterRefresh.accessToken) {
        retryHeaders.Authorization = `Bearer ${tokensAfterRefresh.accessToken}`;
      }

      const retryResp = await fetch(gatewayPath, {
        method: input.method,
        headers: retryHeaders,
        body: input.body === undefined ? undefined : JSON.stringify(input.body),
      });

      if (retryResp.ok) {
        const text = await retryResp.text();
        const json = text ? SafeJson(text) : undefined;
        return camelizeKeysDeep<T>(json);
      }
    } catch {
      // Fall through to normalized 401 error.
    }
  }

  const err = await parseError(resp);
  throw err;
}

export async function apiGet<T>(path: string): Promise<T> {
  return apiRequest<T>({ method: "GET", path });
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>({ method: "POST", path, body });
}

export async function apiPut<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>({ method: "PUT", path, body });
}

export async function apiPatch<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>({ method: "PATCH", path, body });
}

export async function apiDelete<T>(path: string): Promise<T> {
  return apiRequest<T>({ method: "DELETE", path });
}

