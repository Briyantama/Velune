import { cookies } from "next/headers";

export const ACCESS_COOKIE = "velune_access_token";
export const REFRESH_COOKIE = "velune_refresh_token";

type CookieStore = ReturnType<typeof cookies>;
type CookieStoreResolved = CookieStore extends Promise<infer T> ? T : CookieStore;
type CookieSetOptions = Parameters<CookieStoreResolved["set"]>[2];

export function cookieOptions(): CookieSetOptions {
  return {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/"
  };
}
