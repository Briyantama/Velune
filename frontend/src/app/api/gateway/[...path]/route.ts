import { cookies, headers } from "next/headers";
import { NextResponse } from "next/server";
import { ACCESS_COOKIE, REFRESH_COOKIE, cookieOptions } from "@/src/lib/auth/cookies";
import type { BackendError, TokenResponse } from "@/src/lib/api/backend-types";
import { ensureCorrelationId, gatewayFetch } from "@/src/lib/api/http";
import { SafeJson } from "@/src/lib/utils";

export const dynamic = "force-dynamic";

type RouteCtx = { params: { path: string[] } };

export async function GET(req: Request, ctx: RouteCtx) {
  return proxy(req, ctx);
}
export async function POST(req: Request, ctx: RouteCtx) {
  return proxy(req, ctx);
}
export async function PUT(req: Request, ctx: RouteCtx) {
  return proxy(req, ctx);
}
export async function PATCH(req: Request, ctx: RouteCtx) {
  return proxy(req, ctx);
}
export async function DELETE(req: Request, ctx: RouteCtx) {
  return proxy(req, ctx);
}

async function proxy(req: Request, ctx: RouteCtx) {
  const jar = await cookies();
  const inboundHeaders = await headers();
  const cid = ensureCorrelationId(inboundHeaders.get("x-correlation-id"));

  const url = new URL(req.url);
  const path = "/api/v1/" + ctx.params.path.join("/");
  const upstreamPath = path + (url.search ? url.search : "");

  const method = req.method.toUpperCase();
  const contentType = req.headers.get("content-type") ?? "";
  const hasBody = method !== "GET" && method !== "HEAD";
  const body =
    hasBody && contentType.includes("application/json") ? ((await req.json()) as unknown) : undefined;

  const access = jar.get(ACCESS_COOKIE)?.value ?? "";
  const first = await gatewayFetch({
    path: upstreamPath,
    method,
    body,
    accessToken: access,
    correlationId: cid
  });

  if (first.status !== 401) {
    return passthrough(first, cid);
  }

  const refreshed = await tryRefresh(Promise.resolve(jar), cid);
  if (!refreshed) {
    return passthrough(first, cid);
  }

  const nextAccess = jar.get(ACCESS_COOKIE)?.value ?? "";
  const second = await gatewayFetch({
    path: upstreamPath,
    method,
    body,
    accessToken: nextAccess,
    correlationId: cid
  });
  return passthrough(second, cid);
}

async function tryRefresh(jar: ReturnType<typeof cookies>, cid: string): Promise<boolean> {
  const refresh = (await jar).get(REFRESH_COOKIE)?.value ?? "";
  if (!refresh) return false;

  const resp = await gatewayFetch({
    path: "/api/v1/auth/refresh",
    method: "POST",
    body: { refresh_token: refresh },
    correlationId: cid
  });

  if (!resp.ok) {
    return false;
  }

  const tokens = (await resp.json()) as TokenResponse;
  if (!tokens?.access_token || !tokens?.refresh_token) return false;

  (await jar).set(ACCESS_COOKIE, tokens.access_token, {
    ...cookieOptions(),
    maxAge: Math.max(1, tokens.expires_in)
  });
  (await jar).set(REFRESH_COOKIE, tokens.refresh_token, {
    ...cookieOptions(),
    maxAge: 60 * 60 * 24 * 30
  });
  return true;
}

async function passthrough(resp: Response, cid: string) {
  const contentType = resp.headers.get("content-type") ?? "application/json";

  // Handle json and non-json bodies uniformly.
  const raw = await resp.text();
  const init = {
    status: resp.status,
    headers: {
      "Content-Type": contentType,
      "X-Correlation-ID": cid
    }
  };

  if (!raw) return new NextResponse(null, init);

  // Normalize errors to backend `{code,message}` when possible.
  if (resp.status >= 400 && contentType.includes("application/json")) {
    const parsed = SafeJson(raw) as Partial<BackendError> | undefined;
    if (parsed?.code && parsed?.message) {
      return NextResponse.json({ code: parsed.code, message: parsed.message }, init);
    }
  }

  if (contentType.includes("application/json")) {
    const parsed = SafeJson(raw);
    return NextResponse.json(parsed ?? { code: "INVALID_JSON", message: "invalid upstream json" }, init);
  }

  return new NextResponse(raw, init);
}
