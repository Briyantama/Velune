import type { CookieSerializeOptions } from "next/dist/compiled/cookie";

export const ACCESS_COOKIE = "velune_access_token";
export const REFRESH_COOKIE = "velune_refresh_token";

export function cookieOptions(): Pick<CookieSerializeOptions, "httpOnly" | "sameSite" | "secure" | "path"> {
  return {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/"
  };
}

