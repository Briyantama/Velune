"use client";

import * as React from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { PageHeader } from "@/src/components/common/page-header";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { EmptyState } from "@/src/components/common/empty-state";
import { DataTable, type ColumnDef } from "@/src/components/common/data-table";
import { Button } from "@/src/components/ui/button";
import { Input } from "@/src/components/ui/input";
import { Label } from "@/src/components/ui/label";
import { useApiToasts } from "@/src/lib/api/toast";
import { queryKeys } from "@/src/lib/query/keys";
import { adminClient } from "@/src/features/admin/services/adminClient";

type Row = {
  service: string;
  id: string;
  eventType: string;
  status: string;
  retryCount: number;
  nextRetryAt: string;
  createdAt: string;
  payloadPreview: string;
};

export default function AdminOutboxClient() {
  const toast = useApiToasts();
  const qc = useQueryClient();
  const [service, setService] = React.useState<
    "all" | "transaction" | "budget"
  >("all");
  const [status, setStatus] = React.useState("");
  const [limit, setLimit] = React.useState(50);

  const q = useQuery({
    queryKey: queryKeys.adminOutbox({
      service,
      status: status || undefined,
      limit,
    }),
    queryFn: () =>
      adminClient.outbox({ service, status: status || undefined, limit }),
  });

  const retryM = useMutation({
    mutationFn: (input: {
      service: "transaction" | "budget";
      outbox_id: string;
    }) => adminClient.outboxRetry(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["adminOutbox"] }),
  });

  const rows: Row[] = React.useMemo(() => {
    const out: Row[] = [];
    const payload = q.data ?? {};
    for (const [svc, list] of Object.entries(payload)) {
      for (const r of list as any[]) {
        out.push({
          service: svc,
          id: String(r.id),
          eventType: String(r.eventType ?? r.event_type ?? ""),
          status: String(r.status ?? ""),
          retryCount: Number(r.retryCount ?? r.retry_count ?? 0),
          nextRetryAt: String(r.nextRetryAt ?? r.next_retry_at ?? ""),
          createdAt: String(r.createdAt ?? r.created_at ?? ""),
          payloadPreview: String(r.preview ?? r.payload_preview ?? ""),
        });
      }
    }
    return out;
  }, [q.data]);

  const cols: ColumnDef<Row>[] = [
    { key: "svc", header: "Service", cell: (r) => r.service },
    { key: "type", header: "Event", cell: (r) => r.eventType },
    { key: "status", header: "Status", cell: (r) => r.status },
    { key: "retry", header: "Retries", cell: (r) => String(r.retryCount) },
    {
      key: "created",
      header: "Created",
      cell: (r) => (r.createdAt ? new Date(r.createdAt).toLocaleString() : "—"),
    },
    {
      key: "actions",
      header: "",
      className: "w-[1%] whitespace-nowrap",
      cell: (r) => (
        <Button
          variant="outline"
          size="sm"
          disabled={
            retryM.isPending ||
            (r.service !== "transaction" && r.service !== "budget")
          }
          onClick={async () => {
            if (r.service !== "transaction" && r.service !== "budget") return;
            try {
              await retryM.mutateAsync({ service: r.service, outbox_id: r.id });
              toast.showSuccess("Retry scheduled", r.id);
            } catch (e) {
              toast.showError(e, "Retry failed");
            }
          }}
        >
          Retry
        </Button>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        title="Outbox"
        description="Inspect pending/failed outbox rows and retry."
      />

      <div className="mb-4 grid grid-cols-1 gap-3 rounded-2xl border bg-card p-4 shadow-soft md:grid-cols-4">
        <div className="grid gap-2">
          <Label>Service</Label>
          <select
            className="h-10 rounded-md border bg-background px-3 text-sm"
            value={service}
            onChange={(e) => setService(e.target.value as any)}
          >
            <option value="all">all</option>
            <option value="transaction">transaction</option>
            <option value="budget">budget</option>
          </select>
        </div>
        <div className="grid gap-2">
          <Label>Status (optional)</Label>
          <Input
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            placeholder="pending|failed|..."
          />
        </div>
        <div className="grid gap-2">
          <Label>Limit</Label>
          <Input
            type="number"
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
          />
        </div>
        <div className="flex items-end">
          <Button variant="secondary" onClick={() => q.refetch()}>
            Refresh
          </Button>
        </div>
      </div>

      {q.isLoading ? <LoadingSkeleton /> : null}
      {q.isError ? (
        <EmptyState
          title="Couldn’t load outbox"
          description="Ensure admin-service DB URLs are configured."
          actionLabel="Retry"
          onAction={() => q.refetch()}
        />
      ) : null}
      {q.data ? (
        <DataTable
          columns={cols}
          rows={rows}
          rowKey={(r) => `${r.service}:${r.id}`}
        />
      ) : null}
    </div>
  );
}
