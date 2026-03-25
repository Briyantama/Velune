"use client";

import * as React from "react";
import { useQueries, useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { PageHeader } from "@/src/components/common/page-header";
import { StatCard } from "@/src/components/common/stat-card";
import { EmptyState } from "@/src/components/common/empty-state";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { DataTable, type ColumnDef } from "@/src/components/common/data-table";
import { StatusBadge } from "@/src/components/common/status-badge";
import { formatMoneyMinor } from "@/src/lib/money/format";
import { queryKeys } from "@/src/lib/query/keys";
import { metaApi } from "@/src/features/meta/services/metaApi";
import {
  budgetsApi,
  type Budget,
} from "@/src/features/budgets/services/budgetsApi";
import {
  transactionsApi,
  type Transaction,
} from "@/src/features/transactions/services/transactionsApi";

export default function DashboardPageClient() {
  const currency = "USD";
  const now = new Date();
  const from = new Date(now.getTime() - 30 * 24 * 3600 * 1000).toISOString();
  const to = now.toISOString();

  const accountsQ = useQuery({
    queryKey: queryKeys.accounts({ page: 1, limit: 200 }),
    queryFn: () => metaApi.listAccounts({ page: 1, limit: 200 }),
  });

  const summaryQ = useQuery({
    queryKey: queryKeys.transactionSummary({ from, to, currency }),
    queryFn: () => transactionsApi.summary({ from, to, currency }),
  });

  const budgetsQ = useQuery({
    queryKey: queryKeys.budgets({ page: 1, limit: 50 }),
    queryFn: () => budgetsApi.list({ page: 1, limit: 50 }),
  });

  const recentTxQ = useQuery({
    queryKey: queryKeys.transactions({
      page: 1,
      limit: 10,
      from,
      to,
      currency,
    }),
    queryFn: () =>
      transactionsApi.list({ page: 1, limit: 10, from, to, currency }),
  });

  const budgetUsageQs = useQueries({
    queries: (budgetsQ.data?.items ?? []).slice(0, 6).map((b) => ({
      queryKey: queryKeys.budgetUsage(b.id),
      queryFn: () => budgetsApi.usage(b.id),
      enabled: !!budgetsQ.data,
    })),
  });

  if (
    accountsQ.isLoading ||
    summaryQ.isLoading ||
    budgetsQ.isLoading ||
    recentTxQ.isLoading
  ) {
    return (
      <div>
        <PageHeader
          title="Dashboard"
          description="Overview of balances, budgets, and recent activity."
        />
        <LoadingSkeleton />
      </div>
    );
  }

  if (
    accountsQ.isError ||
    summaryQ.isError ||
    budgetsQ.isError ||
    recentTxQ.isError
  ) {
    return (
      <div>
        <PageHeader
          title="Dashboard"
          description="Overview of balances, budgets, and recent activity."
        />
        <EmptyState
          title="Couldn’t load dashboard"
          description="Check gateway, auth, and service health."
        />
      </div>
    );
  }

  const accounts = accountsQ.data?.items ?? [];
  const totalBalanceMinor = accounts.reduce(
    (sum, a) => sum + (a.balanceMinor ?? 0),
    0,
  );

  const cols: ColumnDef<Transaction>[] = [
    {
      key: "date",
      header: "Date",
      cell: (t) => new Date(t.occurredAt).toLocaleDateString(),
    },
    { key: "type", header: "Type", cell: (t) => t.type },
    {
      key: "amount",
      header: "Amount",
      cell: (t) =>
        formatMoneyMinor({ amountMinor: t.amountMinor, currency: t.currency }),
    },
    {
      key: "desc",
      header: "Description",
      cell: (t) => (
        <Link
          className="underline underline-offset-4"
          href={`/transactions/${t.id}`}
        >
          {t.description || "—"}
        </Link>
      ),
    },
  ];

  const overspends = budgetUsageQs
    .map((q) => q.data)
    .filter(Boolean)
    .filter((u) => u!.isOverspent)
    .slice(0, 3);

  return (
    <div>
      <PageHeader
        title="Dashboard"
        description="Overview of balances, budgets, and recent activity."
      />

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <StatCard
          title="Total balance"
          value={formatMoneyMinor({ amountMinor: totalBalanceMinor, currency })}
          hint={`${accounts.length} accounts`}
        />
        <StatCard
          title="Last 30 days net"
          value={formatMoneyMinor({
            amountMinor: summaryQ.data?.netMinor ?? 0,
            currency,
          })}
          hint={
            <span>
              Income{" "}
              {formatMoneyMinor({
                amountMinor: summaryQ.data?.incomeMinor ?? 0,
                currency,
              })}{" "}
              · Expense{" "}
              {formatMoneyMinor({
                amountMinor: summaryQ.data?.expenseMinor ?? 0,
                currency,
              })}
            </span>
          }
        />
        <StatCard
          title="Overspend alerts"
          value={overspends.length ? `${overspends.length}` : "0"}
          hint={
            overspends.length
              ? "Some budgets exceeded their limit."
              : "All good in the sampled budgets."
          }
        />
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="rounded-2xl border bg-card p-6 shadow-soft">
          <div className="flex items-center justify-between">
            <div className="text-base font-semibold">Budget usage (sample)</div>
            <Link
              className="text-sm underline underline-offset-4"
              href="/budgets"
            >
              View all
            </Link>
          </div>
          <div className="mt-4 grid gap-3">
            {(budgetsQ.data?.items ?? []).slice(0, 6).map((b, idx) => (
              <BudgetUsageRow
                key={b.id}
                budget={b}
                usage={budgetUsageQs[idx]?.data}
              />
            ))}
          </div>
        </div>

        <div className="rounded-2xl border bg-card p-6 shadow-soft">
          <div className="flex items-center justify-between">
            <div className="text-base font-semibold">Recent transactions</div>
            <Link
              className="text-sm underline underline-offset-4"
              href="/transactions"
            >
              Open ledger
            </Link>
          </div>
          <div className="mt-4">
            <DataTable
              columns={cols}
              rows={recentTxQ.data?.items ?? []}
              rowKey={(t) => t.id}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function BudgetUsageRow({ budget, usage }: { budget: Budget; usage?: any }) {
  const spent = usage?.spentMinor ?? 0;
  const limit = budget.limitAmountMinor ?? 0;
  const pct = limit > 0 ? Math.min(100, Math.round((spent / limit) * 100)) : 0;
  const overspent = usage?.isOverspent ?? false;
  const statusLabel = overspent ? "Overspent" : "On track";

  return (
    <div className="rounded-xl border bg-background p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold">{budget.name}</div>
          <div className="mt-2 flex items-center gap-2">
            <StatusBadge
              variant={overspent ? "warning" : "success"}
              label={statusLabel}
            />
            <div className="text-xs text-muted-foreground">
              {new Date(budget.startDate).toLocaleDateString()} →{" "}
              {new Date(budget.endDate).toLocaleDateString()}
            </div>
          </div>
        </div>
        <div className="text-right text-xs text-muted-foreground">
          {formatMoneyMinor({ amountMinor: spent, currency: budget.currency })}{" "}
          /{" "}
          {formatMoneyMinor({ amountMinor: limit, currency: budget.currency })}
        </div>
      </div>
      <div className="mt-3 h-2 w-full overflow-hidden rounded-full bg-muted">
        <div
          className={"h-2 " + (overspent ? "bg-warning" : "bg-primary")}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
