"use client";

import { PageHeader } from "@/src/components/common/page-header";
import { EmptyState } from "@/src/components/common/empty-state";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { Card, CardContent } from "@/src/components/ui/card";
import { useNotificationsPing } from "@/src/features/notifications/hooks/useNotificationsPing";

export default function NotificationsPageClient() {
  const q = useNotificationsPing();

  return (
    <div>
      <PageHeader
        title="Notifications"
        description="Alerts for budgets and important events."
      />

      {q.isLoading ? <LoadingSkeleton /> : null}
      {q.isError ? (
        <EmptyState
          title="Notifications service unavailable"
          description="Ping failed. Check service health and auth."
          actionLabel="Retry"
          onAction={() => q.refetch()}
        />
      ) : null}

      {q.data ? (
        <Card>
          <CardContent className="p-6">
            <div className="text-sm">
              <div className="font-semibold">Connectivity</div>
              <div className="mt-1 text-sm text-muted-foreground">
                Ping status: {q.data?.status ?? "unknown"}
              </div>
            </div>
          </CardContent>
        </Card>
      ) : null}

      <div className="mt-6">
        <EmptyState
          title="No notification feed endpoint yet"
          description="The backend currently exposes ping only. When a user notification list endpoint is added, this page will render read/unread, delivery state, and filters."
        />
      </div>
    </div>
  );
}
