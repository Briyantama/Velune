import { camelizeKeysDeep } from "@/src/lib/api/normalize";
import { SafeJson } from "@/src/lib/utils";

type ClientError = { code: string; message: string; status: number };

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const resp = await fetch(`/api/admin/${stripLeadingSlash(path)}`, {
    cache: "no-store",
    ...init,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    }
  });
  const text = await resp.text();
  const json = text ? SafeJson(text) : undefined;
  if (!resp.ok) {
    const err: ClientError = {
      code: (json as any)?.code ?? "HTTP_ERROR",
      message: (json as any)?.message ?? `Request failed (${resp.status})`,
      status: resp.status
    };
    throw err;
  }
  return camelizeKeysDeep<T>(json);
}

function stripLeadingSlash(p: string) {
  return p.startsWith("/") ? p.slice(1) : p;
}

export const adminClient = {
  health() {
    return request<{ services: Record<string, string>; outboxPending: number; dlqMessages: number; notificationFailures: number }>(
      "/internal/admin/health"
    );
  },
  dlqPeek(limit: number) {
    return request<{ messages: any[] }>(`/internal/admin/dlq?limit=${encodeURIComponent(String(limit))}`);
  },
  dlqReplay(input: { event_id: string; target?: string }) {
    return request<{ status: string; event_id: string }>(`/internal/admin/dlq/replay`, {
      method: "POST",
      body: JSON.stringify({ event_id: input.event_id, target: input.target ?? "" })
    });
  },
  outbox(args: { service: "transaction" | "budget" | "all"; status?: string; limit: number }) {
    const p = new URLSearchParams({ service: args.service, limit: String(args.limit) });
    if (args.status) p.set("status", args.status);
    return request<Record<string, any[]>>(`/internal/admin/outbox?${p.toString()}`);
  },
  outboxRetry(input: { service: "transaction" | "budget"; outbox_id: string }) {
    return request<{ status: string; outbox_id: string }>(`/internal/admin/outbox/retry`, {
      method: "POST",
      body: JSON.stringify(input)
    });
  },
  reconcileBalance() {
    return request<any>(`/internal/admin/reconcile/balance`, { method: "POST" });
  },
  reconcileBudget() {
    return request<any>(`/internal/admin/reconcile/budget`, { method: "POST" });
  },
  reconcileLogs(args: { service: "transaction" | "budget" | "all"; type?: string; limit: number }) {
    const p = new URLSearchParams({ service: args.service, limit: String(args.limit) });
    if (args.type) p.set("type", args.type);
    return request<{ logs: any[] }>(`/internal/admin/reconcile/logs?${p.toString()}`);
  },
  eventsReplay(input: { event_type: string; from?: string; to?: string; dry_run?: boolean }) {
    const p = new URLSearchParams();
    if (input.dry_run) p.set("dry_run", "1");
    const qs = p.toString();
    return request<any>(`/internal/admin/events/replay${qs ? `?${qs}` : ""}`, {
      method: "POST",
      body: JSON.stringify({ event_type: input.event_type, from: input.from ?? "", to: input.to ?? "" })
    });
  }
};

