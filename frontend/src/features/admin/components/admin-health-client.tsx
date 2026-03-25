"use client";

import { useQuery } from "@tanstack/react-query";
import { PageHeader } from "@/src/components/common/page-header";
import { EmptyState } from "@/src/components/common/empty-state";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { Card, CardContent } from "@/src/components/ui/card";
import { queryKeys } from "@/src/lib/query/keys";
import { adminClient } from "@/src/features/admin/services/adminClient";

export default function AdminHealthClient() {
  const q = useQuery({
    queryKey: queryKeys.adminHealth(),
    queryFn: () => adminClient.health(),
  });

  return (
    <div>
      <PageHeader
        title="Admin health"
        description="Service probes, outbox backlog, DLQ depth."
      />
      {q.isLoading ? <LoadingSkeleton /> : null}
      {q.isError ? (
        <EmptyState
          title="Couldn’t load admin health"
          description="Ensure ADMIN_SERVICE_URL and ADMIN_API_KEY are configured."
          actionLabel="Retry"
          onAction={() => q.refetch()}
        />
      ) : null}
      {q.data ? (
        <Card>
          <CardContent className="p-6 text-sm">
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <div className="text-xs font-semibold text-muted-foreground">
                  Services
                </div>
                <div className="mt-2 grid gap-2">
                  {Object.entries(q.data.services).map(([k, v]) => (
                    <div
                      key={k}
                      className="flex items-center justify-between rounded-xl border bg-background px-4 py-3"
                    >
                      <div className="font-medium">{k}</div>
                      <div className="text-xs text-muted-foreground">{v}</div>
                    </div>
                  ))}
                </div>
              </div>
              <div className="space-y-3">
                <Metric
                  label="Outbox pending"
                  value={String(q.data.outboxPending)}
                />
                <Metric
                  label="DLQ messages"
                  value={String(q.data.dlqMessages)}
                />
                <Metric
                  label="Notification failures"
                  value={String(q.data.notificationFailures)}
                />
              </div>
            </div>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border bg-background p-4">
      <div className="text-xs font-semibold text-muted-foreground">{label}</div>
      <div className="mt-2 text-lg font-semibold">{value}</div>
    </div>
  );
}
