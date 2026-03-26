"use client";

import { useQuery } from "@tanstack/react-query";
import { PageHeader } from "@/src/components/common/page-header";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { EmptyState } from "@/src/components/common/empty-state";
import { Card, CardContent } from "@/src/components/ui/card";
import { transactionsApi } from "@/src/features/transactions/services/transactionsApi";
import { queryKeys } from "@/src/lib/query/keys";
import { formatMoneyMinor } from "@/src/lib/money/format";

export default function TransactionDetailClient({ id }: { id: string }) {
  const q = useQuery({
    queryKey: queryKeys.transaction(id),
    queryFn: () => transactionsApi.get(id),
  });

  if (q.isLoading) return <LoadingSkeleton />;
  if (q.isError) {
    return (
      <EmptyState
        title="Couldn’t load transaction"
        description="Check auth and try again."
        actionLabel="Retry"
        onAction={() => q.refetch()}
      />
    );
  }

  const t = q.data;
  return (
    <div>
      <PageHeader title="Transaction" description={t?.id ?? "N/A"} />
      <Card>
        <CardContent className="p-6 text-sm">
          <dl className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <KV k="Type" v={t?.type ?? "N/A"} />
            <KV
              k="Amount"
              v={formatMoneyMinor({
                amountMinor: t?.amountMinor ?? 0,
                currency: t?.currency ?? "IDR",
              })}
            />
            <KV
              k="Occurred at"
              v={new Date(t?.occurredAt ?? "").toLocaleString()}
            />
            <KV k="Currency" v={t?.currency ?? "IDR"} />
            <KV k="Account ID" v={t?.accountId ?? "N/A"} />
            <KV k="Category ID" v={t?.categoryId ?? "N/A"} />
            <KV
              k="Counterparty Account ID"
              v={t?.counterpartyAccountId ?? "N/A"}
            />
            <KV k="Version" v={String(t?.version ?? "N/A")} />
            <KV k="Description" v={t?.description ?? "N/A"} wide />
          </dl>
        </CardContent>
      </Card>
    </div>
  );
}

function KV({ k, v, wide }: { k: string; v: string; wide?: boolean }) {
  return (
    <div className={wide ? "md:col-span-2" : ""}>
      <dt className="text-xs font-semibold text-muted-foreground">{k}</dt>
      <dd className="mt-1 break-words">{v}</dd>
    </div>
  );
}
