import { cookies, headers } from "next/headers";
import { NextResponse } from "next/server";
import type { MeResponse, TokenResponse } from "@/src/lib/api/backend-types";
import { ApiError, ensureCorrelationId, gatewayFetch, readJsonOrThrow } from "@/src/lib/api/http";
import { ACCESS_COOKIE, REFRESH_COOKIE, cookieOptions } from "@/src/lib/auth/cookies";
import { serverEnv } from "@/src/lib/env/server";
import { SafeJson } from "@/src/lib/utils";

export const dynamic = "force-dynamic";

type RouteCtx = { params: { path: string[] } };

export async function GET(req: Request, ctx: RouteCtx) {
  return proxy(req, ctx);
}
export async function POST(req: Request, ctx: RouteCtx) {
  return proxy(req, ctx);
}

async function proxy(req: Request, ctx: RouteCtx) {
  if (!serverEnv.adminServiceUrl) {
    return NextResponse.json({ code: "MISCONFIGURED", message: "ADMIN_SERVICE_URL is not set" }, { status: 503 });
  }
  if (!serverEnv.adminApiKey) {
    return NextResponse.json({ code: "MISCONFIGURED", message: "ADMIN_API_KEY is not set" }, { status: 503 });
  }

  const reqHeaders = await headers();
  const cid = ensureCorrelationId(reqHeaders.get("x-correlation-id"));

  const jar = await cookies();
  const access = jar.get(ACCESS_COOKIE)?.value ?? "";
  const refresh = jar.get(REFRESH_COOKIE)?.value ?? "";

  if (!access || !refresh) {
    return NextResponse.json(
      { code: "AUTH_REQUIRED", message: "authentication required" },
      { status: 401, headers: { "X-Correlation-ID": cid } },
    );
  }

  // Admin service uses API-key auth; enforce user session validity here so expired
  // sessions are cleared + users are redirected.
  try {
    const meResp = await gatewayFetch({
      path: "/api/v1/auth/me",
      method: "GET",
      accessToken: access,
      correlationId: cid,
      cache: "no-store",
    });
    await readJsonOrThrow<MeResponse>(meResp);
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      const refreshResp = await gatewayFetch({
        path: "/api/v1/auth/refresh",
        method: "POST",
        body: { refresh_token: refresh },
        correlationId: cid,
        cache: "no-store",
      });

      try {
        const tokens = await readJsonOrThrow<TokenResponse>(refreshResp);
        jar.set(ACCESS_COOKIE, tokens.access_token, {
          ...cookieOptions(),
          maxAge: Math.max(1, tokens.expires_in),
        });
        jar.set(REFRESH_COOKIE, tokens.refresh_token, {
          ...cookieOptions(),
          maxAge: 60 * 60 * 24 * 30,
        });
      } catch {
        jar.set(ACCESS_COOKIE, "", { ...cookieOptions(), maxAge: 0 });
        jar.set(REFRESH_COOKIE, "", { ...cookieOptions(), maxAge: 0 });
        return NextResponse.json(
          { code: "SESSION_EXPIRED", message: "session expired" },
          { status: 401, headers: { "X-Correlation-ID": cid } },
        );
      }
    } else {
      return NextResponse.json(
        { code: "AUTH_VALIDATION_FAILED", message: "authentication validation failed" },
        { status: 401, headers: { "X-Correlation-ID": cid } },
      );
    }
  }

  const url = new URL(req.url);
  const upstream = new URL(serverEnv.adminServiceUrl + "/" + ctx.params.path.join("/"));
  upstream.search = url.search;

  const method = req.method.toUpperCase();
  const contentType = req.headers.get("content-type") ?? "";
  const hasBody = method !== "GET" && method !== "HEAD";
  const body = hasBody && contentType.includes("application/json") ? await req.text() : undefined;

  const resp = await fetch(upstream, {
    method,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      "X-Admin-Key": serverEnv.adminApiKey,
      "X-Correlation-ID": cid
    },
    body: body && body.length ? body : undefined,
    cache: "no-store"
  });

  const raw = await resp.text();
  const init = {
    status: resp.status,
    headers: {
      "Content-Type": resp.headers.get("content-type") ?? "application/json",
      "X-Correlation-ID": cid
    }
  };

  if (!raw) return new NextResponse(null, init);

  if ((init.headers["Content-Type"] ?? "").includes("application/json")) {
    return NextResponse.json(SafeJson(raw) ?? { code: "INVALID_JSON", message: "invalid upstream json" }, init);
  }
  return new NextResponse(raw, init);
}
