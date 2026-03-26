"use client";

import { useState } from "react";
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

type DlqMsg = any;
type DlqRow = { id: string; msg: DlqMsg };

export default function AdminDLQClient() {
  const toast = useApiToasts();
  const qc = useQueryClient();
  const [limit, setLimit] = useState(50);
  const [eventId, setEventId] = useState("");

  const q = useQuery({
    queryKey: queryKeys.adminDlq({ limit }),
    queryFn: () => adminClient.dlqPeek(limit),
  });

  const replayM = useMutation({
    mutationFn: (input: { event_id: string }) => adminClient.dlqReplay(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["adminDlq"] }),
  });

  const cols: ColumnDef<DlqRow>[] = [
    {
      key: "eventId",
      header: "Event ID",
      cell: (r) => String(r.msg.event_id ?? r.msg.eventId ?? "—"),
    },
    {
      key: "type",
      header: "Type",
      cell: (r) => String(r.msg.event_type ?? r.msg.eventType ?? "—"),
    },
    {
      key: "reason",
      header: "Reason",
      cell: (r) => String(r.msg.reason ?? r.msg.error ?? "—"),
    },
  ];

  return (
    <div>
      <PageHeader
        title="DLQ"
        description="Peek messages and replay by event_id."
      />

      <div className="mb-4 grid grid-cols-1 gap-3 rounded-2xl border bg-card p-4 shadow-soft md:grid-cols-3">
        <div className="grid gap-2">
          <Label>Limit</Label>
          <Input
            type="number"
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
          />
        </div>
        <div className="grid gap-2 md:col-span-2">
          <Label>Replay event_id</Label>
          <div className="flex gap-2">
            <Input
              value={eventId}
              onChange={(e) => setEventId(e.target.value)}
              placeholder="uuid"
            />
            <Button
              disabled={!eventId.trim() || replayM.isPending}
              onClick={async () => {
                try {
                  await replayM.mutateAsync({ event_id: eventId.trim() });
                  toast.showSuccess("Replay scheduled", eventId.trim());
                  setEventId("");
                } catch (e) {
                  toast.showError(e, "Replay failed");
                }
              }}
            >
              Replay
            </Button>
          </div>
        </div>
      </div>

      {q.isLoading ? <LoadingSkeleton /> : null}
      {q.isError ? (
        <EmptyState
          title="Couldn’t load DLQ"
          description="Ensure broker + admin-service are configured."
          actionLabel="Retry"
          onAction={() => q.refetch()}
        />
      ) : null}
      {q.data ? (
        <DataTable
          columns={cols}
          rows={(q.data.messages ?? []).map((m, idx) => ({
            id: String(idx),
            msg: m,
          }))}
          rowKey={(r) => r.id}
        />
      ) : null}
    </div>
  );
}
