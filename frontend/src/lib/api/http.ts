import { serverEnv } from "@/src/lib/env/server";
import type { BackendError } from "@/src/lib/api/backend-types";
import { SafeJson } from "../utils";

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

async function proxy(req: Request, ctx: RouteCtx): Promise<Response> {
  const url = new URL(req.url);
  const path = "/api/v1/" + ctx.params.path.join("/");
  const upstreamPath = path + (url.search ? url.search : "");

  const method = req.method.toUpperCase();
  const contentType = req.headers.get("content-type") ?? "";
  const hasBody = method !== "GET" && method !== "HEAD" && method !== "DELETE";

  const body =
    hasBody && contentType.includes("application/json") ? ((await req.json()) as unknown) : undefined;

  const authorization = req.headers.get("authorization") ?? "";
  const accessToken = authorization.startsWith("Bearer ") ? authorization.slice("Bearer ".length) : undefined;
  const cid = ensureCorrelationId(req.headers.get("x-correlation-id"));

  return gatewayFetch({
    path: upstreamPath,
    method,
    accessToken,
    correlationId: cid,
    body
  });
}

export class ApiError extends Error {
  code: string;
  status: number;

  constructor(init: { code: string; message: string; status: number }) {
    super(init.message);
    this.code = init.code;
    this.status = init.status;
  }
}

export function ensureCorrelationId(inbound?: string | null): string {
  const v = (inbound ?? "").trim();
  return v || crypto.randomUUID();
}

export async function gatewayFetch(input: {
  path: string;
  method: string;
  headers?: Record<string, string>;
  body?: unknown;
  accessToken?: string;
  correlationId?: string;
  cache?: RequestCache;
}): Promise<Response> {
  const url = new URL(serverEnv.gatewayBaseUrl + input.path);
  const headers = new Headers(input.headers ?? {});
  headers.set("Accept", "application/json");
  headers.set("Content-Type", "application/json");
  if (input.correlationId) headers.set("X-Correlation-ID", input.correlationId);
  if (input.accessToken) headers.set("Authorization", `Bearer ${input.accessToken}`);

  return fetch(url, {
    method: input.method,
    headers,
    body: input.body === undefined ? undefined : JSON.stringify(input.body),
    cache: input.cache ?? "no-store"
  });
}

export async function readJsonOrThrow<T>(resp: Response): Promise<T> {
  const text = await resp.text();
  const json = text ? SafeJson(text) : undefined;

  if (resp.ok) {
    // Backend responses are standardized as `{ data: ... }` envelopes.
    const payload = (json as any)?.data ?? json;
    return payload as T;
  }

  const be = (json ?? {}) as Partial<BackendError>;
  const code =
    typeof be.code === "string" && be.code
      ? be.code
      : `HTTP_${resp.status}`;
  const message =
    (typeof be.error === "string" && be.error && be.error) ||
    (typeof be.message === "string" && be.message && be.message) ||
    `Request failed (${resp.status})`;
  throw new ApiError({ code, message, status: resp.status });
}
