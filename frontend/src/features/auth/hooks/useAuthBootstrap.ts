"use client";

import { useEffect } from "react";
import { camelizeKeysDeep } from "@/src/lib/api/normalize";
import { SafeJson } from "@/src/lib/utils";
import { authClient, type Me as MeFromApi } from "@/src/features/auth/services/authClient";
import {
  clearAllLocalSessionArtifacts,
  getStoredTokensForMode,
  saveLocalTokens,
  setSessionExpiresAtCookie
} from "@/src/services/authStorage";
import { useAppDispatch } from "@/src/store/hooks";
import { clearAuthState, setStatus } from "@/src/store/slices/authSlice";
import { clearMe, setMe } from "@/src/store/slices/userSessionSlice";
import { type StorageMode } from "@/src/store/slices/consentSlice";

async function fetchMe(storageMode: StorageMode, accessToken: string | null): Promise<MeFromApi> {
  const headers: Record<string, string> = { Accept: "application/json" };
  if (storageMode === "localStorage" && accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }

  const resp = await fetch("/api/auth/me", { method: "GET", headers });
  const text = await resp.text();
  const json = text ? SafeJson(text) : undefined;

  if (!resp.ok) {
    const code = (json as { code?: string })?.code ?? "AUTH_ME_FAILED";
    const message = (json as { message?: string })?.message ?? `Request failed (${resp.status})`;
    throw { code, message, status: resp.status };
  }

  // Server returns snake_case; normalize to client shape.
  return camelizeKeysDeep<MeFromApi>(json);
}

export function useAuthBootstrap(storageMode: StorageMode | null) {
  const dispatch = useAppDispatch();

  useEffect(() => {
    if (!storageMode) return;

    let cancelled = false;

    const bootstrap = async () => {
      dispatch(setStatus("hydrating"));
      try {
        if (storageMode === "cookie") {
          const me = await authClient.me();
          if (cancelled) return;
          dispatch(setMe(me));
          dispatch(setStatus("authenticated"));
          return;
        }

        const { accessToken } = getStoredTokensForMode("localStorage");
        if (!accessToken) {
          dispatch(clearMe());
          dispatch(setStatus("unauthenticated"));
          return;
        }

        const me = await fetchMe("localStorage", accessToken);
        if (cancelled) return;
        dispatch(setMe(me));
        dispatch(setStatus("authenticated"));
      } catch (e) {
        if (cancelled) return;

        // localStorage mode: attempt refresh once if we have a refresh token.
        const status = typeof (e as { status?: unknown } | undefined)?.status === "number" ? (e as { status: number }).status : undefined;
        if (storageMode === "localStorage" && status === 401) {
          const { refreshToken } = getStoredTokensForMode("localStorage");
          if (refreshToken) {
            try {
              const refreshResp = await fetch("/api/auth/refresh", {
                method: "POST",
                headers: {
                  "x-velune-refresh-token": refreshToken,
                  Accept: "application/json",
                },
              });
              if (!refreshResp.ok) throw new Error("refresh failed");
              const tokens = (await refreshResp.json()) as {
                access_token: string;
                refresh_token: string;
                expires_in: number;
              };

              saveLocalTokens(tokens);
              setSessionExpiresAtCookie(tokens.expires_in);

              const me = await fetchMe("localStorage", tokens.access_token);
              if (cancelled) return;
              dispatch(setMe(me));
              dispatch(setStatus("authenticated"));
              return;
            } catch {
              // Fall through to full logout.
            }
          }
        }

        // Session expired: clear local session artifacts (cookie mode clearance is handled by server).
        clearAllLocalSessionArtifacts();
        dispatch(clearMe());
        dispatch(clearAuthState());
      }
    };

    void bootstrap();

    return () => {
      cancelled = true;
    };
  }, [dispatch, storageMode]);
}

