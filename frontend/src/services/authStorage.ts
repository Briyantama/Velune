import type { TokenResponse } from "@/src/lib/api/backend-types";
import type { StorageMode } from "@/src/store/slices/consentSlice";

export const CONSENT_MODE_COOKIE = "velune_storage_mode";
export const SESSION_EXPIRES_AT_COOKIE = "velune_session_expires_at";

export const LOCAL_ACCESS_TOKEN_KEY = "velune_access_token";
export const LOCAL_REFRESH_TOKEN_KEY = "velune_refresh_token";

function parseCookieValue(rawCookie: string, key: string): string | null {
  const parts = rawCookie.split(";").map((p) => p.trim());
  for (const part of parts) {
    const [k, ...rest] = part.split("=");
    if (k === key) return rest.join("=").trim() || null;
  }
  return null;
}

export function getConsentModeFromCookieString(rawCookie: string): StorageMode | null {
  const v = parseCookieValue(rawCookie, CONSENT_MODE_COOKIE);
  if (v === "cookie" || v === "localStorage") return v;
  return null;
}

type CookieJarLike = {
  get: (key: string) => { value?: string } | undefined;
};

export function getConsentModeFromJar(jar: CookieJarLike): StorageMode | null {
  const v = jar.get(CONSENT_MODE_COOKIE)?.value ?? "";
  if (v === "cookie" || v === "localStorage") return v;
  return null;
}

export function getSessionExpiresAtFromJar(jar: CookieJarLike): number | null {
  const v = jar.get(SESSION_EXPIRES_AT_COOKIE)?.value ?? "";
  if (!v) return null;
  const n = Number(v);
  return Number.isFinite(n) ? n : null;
}

function cookieStringExpiresAtMsToDate(expiresAtMs: number) {
  return new Date(expiresAtMs);
}

function toCookieExpires(date: Date) {
  // Example: "Wed, 25 Mar 2026 07:00:00 GMT"
  return date.toUTCString();
}

function isBrowser() {
  return typeof window !== "undefined";
}

export function getStoredTokensForMode(mode: StorageMode) {
  if (!isBrowser() || mode !== "localStorage") return { accessToken: null as string | null, refreshToken: null as string | null };

  const accessToken = window.localStorage.getItem(LOCAL_ACCESS_TOKEN_KEY);
  const refreshToken = window.localStorage.getItem(LOCAL_REFRESH_TOKEN_KEY);
  return { accessToken, refreshToken };
}

export function saveLocalTokens(tokens: TokenResponse) {
  if (!isBrowser()) return;
  window.localStorage.setItem(LOCAL_ACCESS_TOKEN_KEY, tokens.access_token);
  window.localStorage.setItem(LOCAL_REFRESH_TOKEN_KEY, tokens.refresh_token);
}

export function clearLocalTokens() {
  if (!isBrowser()) return;
  window.localStorage.removeItem(LOCAL_ACCESS_TOKEN_KEY);
  window.localStorage.removeItem(LOCAL_REFRESH_TOKEN_KEY);
  window.localStorage.removeItem("velune_access_token_expires_in");
}

export function setSessionExpiresAtCookie(expiresInSeconds: number) {
  if (!isBrowser()) return;
  // Minimal, non-HttpOnly metadata for middleware gating.
  // `expires_in` is access-token TTL; refresh-token TTL is longer.
  // Your server currently sets refresh cookies to 30 days, so we align middleware gating to that window.
  const refreshTtlSeconds = 60 * 60 * 24 * 30;
  const effectiveTtlSeconds = Math.max(0, expiresInSeconds, refreshTtlSeconds);
  const expiresAtMs = Date.now() + effectiveTtlSeconds * 1000;
  const date = cookieStringExpiresAtMsToDate(expiresAtMs);
  const secure = process.env.NODE_ENV === "production" ? "; Secure" : "";

  document.cookie = `${SESSION_EXPIRES_AT_COOKIE}=${expiresAtMs}; Path=/; Max-Age=${60 * 60 * 24 * 30}${secure}; SameSite=Lax; Expires=${toCookieExpires(
    date,
  )}`;
}

export function clearSessionExpiresAtCookie() {
  if (!isBrowser()) return;
  const secure = process.env.NODE_ENV === "production" ? "; Secure" : "";
  document.cookie = `${SESSION_EXPIRES_AT_COOKIE}=; Path=/; Max-Age=0${secure}; SameSite=Lax`;
}

export function setConsentModeCookie(mode: StorageMode) {
  if (!isBrowser()) return;
  const secure = process.env.NODE_ENV === "production" ? "; Secure" : "";
  // 1 year persistence; consent reset can clear it.
  const maxAgeSeconds = 60 * 60 * 24 * 365;
  document.cookie = `${CONSENT_MODE_COOKIE}=${mode}; Path=/; Max-Age=${maxAgeSeconds}${secure}; SameSite=Lax`;
}

export function clearConsentModeCookie() {
  if (!isBrowser()) return;
  const secure = process.env.NODE_ENV === "production" ? "; Secure" : "";
  document.cookie = `${CONSENT_MODE_COOKIE}=; Path=/; Max-Age=0${secure}; SameSite=Lax`;
}

export function getConsentModeClient(): StorageMode | null {
  if (!isBrowser()) return null;
  return getConsentModeFromCookieString(document.cookie);
}

export function clearAllLocalSessionArtifacts() {
  clearLocalTokens();
  clearSessionExpiresAtCookie();
}

