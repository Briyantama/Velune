import { cookies, headers } from "next/headers";
import { NextResponse } from "next/server";
import { ACCESS_COOKIE, REFRESH_COOKIE, cookieOptions } from "@/src/lib/auth/cookies";
import type { TokenResponse } from "@/src/lib/api/backend-types";
import { ensureCorrelationId, gatewayFetch, readJsonOrThrow } from "@/src/lib/api/http";
import { getConsentModeFromJar } from "@/src/services/authStorage";

export async function POST() {
  const jar = await cookies();
  const reqHeaders = await headers();
  const storageMode = getConsentModeFromJar(jar) ?? "cookie";
  const refreshCookie = jar.get(REFRESH_COOKIE)?.value ?? "";
  const refreshHeader = reqHeaders.get("x-velune-refresh-token") ?? "";
  const cid = ensureCorrelationId(reqHeaders.get("x-correlation-id"));

  if (storageMode === "localStorage") {
    if (!refreshHeader) {
      return NextResponse.json(
        { code: "AUTH_REQUIRED", message: "authentication required" },
        { status: 401, headers: { "X-Correlation-ID": cid } },
      );
    }

    const resp = await gatewayFetch({
      path: "/api/v1/auth/refresh",
      method: "POST",
      body: { refresh_token: refreshHeader },
      correlationId: cid
    });

    const tokens = await readJsonOrThrow<TokenResponse>(resp);
    return NextResponse.json(tokens, { headers: { "X-Correlation-ID": cid } });
  }

  const refresh = refreshCookie;
  if (!refresh) {
    return NextResponse.json(
      { code: "AUTH_REQUIRED", message: "authentication required" },
      { status: 401, headers: { "X-Correlation-ID": cid } },
    );
  }

  const resp = await gatewayFetch({
    path: "/api/v1/auth/refresh",
    method: "POST",
    body: { refresh_token: refresh },
    correlationId: cid
  });

  const tokens = await readJsonOrThrow<TokenResponse>(resp);
  jar.set(ACCESS_COOKIE, tokens.access_token, {
    ...cookieOptions(),
    maxAge: Math.max(1, tokens.expires_in)
  });
  jar.set(REFRESH_COOKIE, tokens.refresh_token, {
    ...cookieOptions(),
    maxAge: 60 * 60 * 24 * 30
  });

  return NextResponse.json({ status: "ok" }, { headers: { "X-Correlation-ID": cid } });
}

