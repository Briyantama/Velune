"use client";

import { useState, useMemo } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { PageHeader } from "@/src/components/common/page-header";
import { EmptyState } from "@/src/components/common/empty-state";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { DataTable, type ColumnDef } from "@/src/components/common/data-table";
import { Button } from "@/src/components/ui/button";
import { Input } from "@/src/components/ui/input";
import { Label } from "@/src/components/ui/label";
import { useApiToasts } from "@/src/lib/api/toast";
import { queryKeys } from "@/src/lib/query/keys";
import { adminClient } from "@/src/features/admin/services/adminClient";

type LogRow = {
  service: string;
  id: string;
  type: string;
  status: string;
  createdAt: string;
  details: unknown;
};

export default function AdminReconciliationClient() {
  const toast = useApiToasts();
  const [service, setService] = useState<
    "all" | "transaction" | "budget"
  >("all");
  const [type, setType] = useState("");
  const [limit, setLimit] = useState(50);

  const logsQ = useQuery({
    queryKey: queryKeys.adminReconcileLogs({
      service,
      type: type || undefined,
      limit,
    }),
    queryFn: () =>
      adminClient.reconcileLogs({ service, type: type || undefined, limit }),
  });

  const balM = useMutation({
    mutationFn: () => adminClient.reconcileBalance(),
    onSuccess: () => logsQ.refetch(),
  });
  const budM = useMutation({
    mutationFn: () => adminClient.reconcileBudget(),
    onSuccess: () => logsQ.refetch(),
  });

  const rows: LogRow[] = useMemo(() => {
    const list = logsQ.data?.logs ?? [];
    return list.map((r: any) => ({
      service: String(r.service),
      id: String(r.id),
      type: String(r.type),
      status: String(r.status),
      createdAt: String(r.createdAt),
      details: r.details,
    }));
  }, [logsQ.data]);

  const cols: ColumnDef<LogRow>[] = [
    { key: "svc", header: "Service", cell: (r) => r.service },
    { key: "type", header: "Type", cell: (r) => r.type },
    { key: "status", header: "Status", cell: (r) => r.status },
    {
      key: "created",
      header: "Created",
      cell: (r) => (r.createdAt ? new Date(r.createdAt).toLocaleString() : "—"),
    },
  ];

  return (
    <div>
      <PageHeader
        title="Reconciliation"
        description="Trigger balance/budget reconciliation and inspect audit logs."
        actions={
          <>
            <Button
              variant="outline"
              disabled={balM.isPending}
              onClick={async () => {
                try {
                  const res = await balM.mutateAsync();
                  toast.showSuccess(
                    "Balance reconcile triggered",
                    JSON.stringify(res),
                  );
                } catch (e) {
                  toast.showError(e, "Reconcile failed");
                }
              }}
            >
              Reconcile balance
            </Button>
            <Button
              variant="outline"
              disabled={budM.isPending}
              onClick={async () => {
                try {
                  const res = await budM.mutateAsync();
                  toast.showSuccess(
                    "Budget reconcile triggered",
                    JSON.stringify(res),
                  );
                } catch (e) {
                  toast.showError(e, "Reconcile failed");
                }
              }}
            >
              Reconcile budget
            </Button>
          </>
        }
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
          <Label>Type (optional)</Label>
          <Input
            value={type}
            onChange={(e) => setType(e.target.value)}
            placeholder="RECONCILE_* etc"
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
          <Button variant="secondary" onClick={() => logsQ.refetch()}>
            Refresh
          </Button>
        </div>
      </div>

      {logsQ.isLoading ? <LoadingSkeleton /> : null}
      {logsQ.isError ? (
        <EmptyState
          title="Couldn’t load logs"
          description="Ensure DB URLs are configured on admin-service."
          actionLabel="Retry"
          onAction={() => logsQ.refetch()}
        />
      ) : null}
      {logsQ.data ? (
        <DataTable
          columns={cols}
          rows={rows}
          rowKey={(r) => `${r.service}:${r.id}`}
        />
      ) : null}
    </div>
  );
}
