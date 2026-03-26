"use client";

import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
} from "recharts";
import { PageHeader } from "@/src/components/common/page-header";
import { FilterBar } from "@/src/components/common/filter-bar";
import { EmptyState } from "@/src/components/common/empty-state";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { Card, CardContent } from "@/src/components/ui/card";
import { Input } from "@/src/components/ui/input";
import { Label } from "@/src/components/ui/label";
import { formatMoneyMinor } from "@/src/lib/money/format";
import { useMonthlyReport } from "@/src/features/reports/hooks/useMonthlyReport";
import { useMemo, useState } from "react";

export default function ReportsPageClient() {
  const now = new Date();
  const [year, setYear] = useState<number>(now.getUTCFullYear());
  const [month, setMonth] = useState<number>(now.getUTCMonth() + 1);
  const [currency, setCurrency] = useState<string>("IDR");

  const q = useMonthlyReport({ year, month, currency });

  const chartData = useMemo(
    () => [
      { name: "Income", valueMinor: q.data?.incomeMinor ?? 0 },
      { name: "Expense", valueMinor: q.data?.expenseMinor ?? 0 },
    ],
    [q.data],
  );

  return (
    <div>
      <PageHeader
        title="Reports"
        description="Monthly summaries and category breakdowns."
      />

      <FilterBar>
        <div className="grid gap-2">
          <Label>Year</Label>
          <Input
            type="number"
            value={year}
            onChange={(e) => setYear(Number(e.target.value))}
          />
        </div>
        <div className="grid gap-2">
          <Label>Month</Label>
          <Input
            type="number"
            value={month}
            onChange={(e) => setMonth(Number(e.target.value))}
            min={1}
            max={12}
          />
        </div>
        <div className="grid gap-2">
          <Label>Currency</Label>
          <Input
            value={currency}
            onChange={(e) => setCurrency(e.target.value)}
            placeholder="IDR"
          />
        </div>
      </FilterBar>

      {q.isLoading ? <LoadingSkeleton /> : null}
      {q.isError ? (
        <EmptyState
          title="Couldn’t load report"
          description="Check gateway and auth."
          actionLabel="Retry"
          onAction={() => q.refetch()}
        />
      ) : null}

      {q.data ? (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <Card>
            <CardContent className="p-6">
              <div className="text-base font-semibold">Income vs expense</div>
              <div className="mt-1 text-sm text-muted-foreground">
                Generated {new Date(q.data?.generatedAt ?? "").toLocaleString()}
              </div>
              <div className="mt-6 h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={chartData}>
                    <XAxis dataKey="name" />
                    <YAxis tickFormatter={(v) => String(v)} />
                    <Tooltip
                      formatter={(v) =>
                        formatMoneyMinor({ amountMinor: Number(v), currency })
                      }
                      labelFormatter={(l) => String(l)}
                    />
                    <Bar
                      dataKey="valueMinor"
                      fill="hsl(var(--primary))"
                      radius={[6, 6, 0, 0]}
                    />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="text-base font-semibold">Category breakdown</div>
              <div className="mt-4 space-y-2">
                {(q.data?.byCategory ?? []).length === 0 ? (
                  <div className="text-sm text-muted-foreground">
                    No category breakdown available.
                  </div>
                ) : (
                  (q.data?.byCategory ?? []).slice(0, 12).map((c) => (
                    <div
                      key={c.categoryId ?? c.categoryName}
                      className="flex items-center justify-between gap-3 rounded-xl border bg-background px-4 py-3"
                    >
                      <div className="min-w-0 truncate text-sm font-medium">
                        {c.categoryName}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {formatMoneyMinor({
                          amountMinor: c.totalMinor ?? 0,
                          currency: c.currency ?? "IDR",
                        })}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      ) : null}
    </div>
  );
}
