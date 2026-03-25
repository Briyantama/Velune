import { cookies, headers } from "next/headers";
import { NextResponse } from "next/server";
import { ACCESS_COOKIE, REFRESH_COOKIE, cookieOptions } from "@/src/lib/auth/cookies";
import type { TokenResponse } from "@/src/lib/api/backend-types";
import { ensureCorrelationId, gatewayFetch, readJsonOrThrow } from "@/src/lib/api/http";

export async function POST(req: Request) {
  const body = (await req.json()) as { email: string; password: string; baseCurrency: string };
  const cid = ensureCorrelationId(headers().get("x-correlation-id"));

  const resp = await gatewayFetch({
    path: "/api/v1/auth/register",
    method: "POST",
    body,
    correlationId: cid
  });

  const tokens = await readJsonOrThrow<TokenResponse>(resp);
  const jar = cookies();
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

