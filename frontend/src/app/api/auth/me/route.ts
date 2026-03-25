import { cookies, headers } from "next/headers";
import { NextResponse } from "next/server";
import { ACCESS_COOKIE, REFRESH_COOKIE, cookieOptions } from "@/src/lib/auth/cookies";
import type { MeResponse, TokenResponse } from "@/src/lib/api/backend-types";
import { ApiError, ensureCorrelationId, gatewayFetch, readJsonOrThrow } from "@/src/lib/api/http";

export async function GET() {
  const jar = await cookies();
  const access = jar.get(ACCESS_COOKIE)?.value ?? "";
  const refresh = jar.get(REFRESH_COOKIE)?.value ?? "";

  const reqHeaders = await headers();
  const cid = ensureCorrelationId(reqHeaders.get("x-correlation-id"));

  if (!access || !refresh) {
    return NextResponse.json(
      { code: "AUTH_REQUIRED", message: "authentication required" },
      { status: 401, headers: { "X-Correlation-ID": cid } },
    );
  }

  const validate = async (accessToken: string) => {
    const resp = await gatewayFetch({
      path: "/api/v1/auth/me",
      method: "GET",
      accessToken,
      correlationId: cid,
      cache: "no-store",
    });
    return readJsonOrThrow<MeResponse>(resp);
  };

  try {
    const me = await validate(access);
    return NextResponse.json(me, { headers: { "X-Correlation-ID": cid } });
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      try {
        const refreshResp = await gatewayFetch({
          path: "/api/v1/auth/refresh",
          method: "POST",
          body: { refresh_token: refresh },
          correlationId: cid,
          cache: "no-store",
        });
        const tokens = await readJsonOrThrow<TokenResponse>(refreshResp);

        jar.set(ACCESS_COOKIE, tokens.access_token, {
          ...cookieOptions(),
          maxAge: Math.max(1, tokens.expires_in),
        });
        jar.set(REFRESH_COOKIE, tokens.refresh_token, {
          ...cookieOptions(),
          maxAge: 60 * 60 * 24 * 30,
        });

        const me = await validate(tokens.access_token);
        return NextResponse.json(me, { headers: { "X-Correlation-ID": cid } });
      } catch {
        jar.set(ACCESS_COOKIE, "", { ...cookieOptions(), maxAge: 0 });
        jar.set(REFRESH_COOKIE, "", { ...cookieOptions(), maxAge: 0 });
        return NextResponse.json(
          { code: "SESSION_EXPIRED", message: "session expired" },
          { status: 401, headers: { "X-Correlation-ID": cid } },
        );
      }
    }

    return NextResponse.json(
      { code: "AUTH_VALIDATION_FAILED", message: "authentication validation failed" },
      { status: 401, headers: { "X-Correlation-ID": cid } },
    );
  }
}

