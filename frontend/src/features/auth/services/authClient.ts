import { camelizeKeysDeep } from "@/src/lib/api/normalize";
import { SafeJson } from "@/src/lib/utils";
import {
  clearAllLocalSessionArtifacts,
  getConsentModeClient,
  getStoredTokensForMode,
  saveLocalTokens,
  setSessionExpiresAtCookie,
} from "@/src/services/authStorage";
import type { TokenResponse } from "@/src/lib/api/backend-types";

export type Me = {
  userId: string;
  email: string;
  baseCurrency: string;
};

export const authClient = {
  async me(): Promise<Me> {
    const storageMode = getConsentModeClient() ?? "cookie";
    const isLocal = storageMode === "localStorage";
    const tokens = isLocal ? getStoredTokensForMode("localStorage") : { accessToken: null, refreshToken: null };

    const headers: Record<string, string> = { Accept: "application/json" };
    if (isLocal && tokens.accessToken) headers.Authorization = `Bearer ${tokens.accessToken}`;

    const resp = await fetch("/api/auth/me", { method: "GET", headers });
    const text = await resp.text();
    const json = text ? SafeJson(text) : undefined;
    if (!resp.ok) {
      const code = (json as { code?: string })?.code ?? "HTTP_ERROR";
      const message = (json as { message?: string })?.message ?? `Request failed (${resp.status})`;

      if (isLocal && resp.status === 401 && tokens.refreshToken) {
        const refreshResp = await fetch("/api/auth/refresh", {
          method: "POST",
          headers: {
            Accept: "application/json",
            "x-velune-refresh-token": tokens.refreshToken,
          },
        });

        if (refreshResp.ok) {
          const refreshedTokens = (await refreshResp.json()) as TokenResponse;
          saveLocalTokens(refreshedTokens);
          setSessionExpiresAtCookie(refreshedTokens.expires_in);

          const retry = await fetch("/api/auth/me", {
            method: "GET",
            headers: {
              Accept: "application/json",
              Authorization: `Bearer ${refreshedTokens.access_token}`,
            },
          });
          const retryText = await retry.text();
          const retryJson = retryText ? SafeJson(retryText) : undefined;
          if (retry.ok) return camelizeKeysDeep<Me>(retryJson);
        }
      }

      throw { code, message, status: resp.status };
    }
    return camelizeKeysDeep<Me>(json);
  },

  async logout(): Promise<void> {
    const storageMode = getConsentModeClient() ?? "cookie";
    if (storageMode === "localStorage") {
      clearAllLocalSessionArtifacts();
    }

    const resp = await fetch("/api/auth/logout", { method: "POST" });
    if (!resp.ok) {
      throw { code: "LOGOUT_FAILED", message: "logout failed", status: resp.status };
    }
  }
};
