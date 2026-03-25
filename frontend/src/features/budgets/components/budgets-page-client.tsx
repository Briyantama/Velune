"use client";

import * as React from "react";
import Link from "next/link";
import { Resolver, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { PageHeader } from "@/src/components/common/page-header";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { EmptyState } from "@/src/components/common/empty-state";
import { DataTable, type ColumnDef } from "@/src/components/common/data-table";
import { ConfirmDialog } from "@/src/components/common/confirm-dialog";
import { Button } from "@/src/components/ui/button";
import { Input } from "@/src/components/ui/input";
import { Label } from "@/src/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/src/components/ui/dialog";
import { StatusBadge } from "@/src/components/common/status-badge";
import { formatMoneyMinor } from "@/src/lib/money/format";
import { useApiToasts } from "@/src/lib/api/toast";
import { budgetUpsertSchema } from "@/src/features/budgets/schema/budgetSchemas";
import {
  useBudgetUsage,
  useBudgetsList,
  useCreateBudget,
  useDeleteBudget,
  useUpdateBudget,
} from "@/src/features/budgets/hooks/useBudgets";
import type {
  Budget,
  BudgetPeriodType,
} from "@/src/features/budgets/services/budgetsApi";
import { FieldError } from "@/src/lib/utils";

type FormValues = {
  name: string;
  periodType: BudgetPeriodType;
  categoryId?: string | null;
  startDate: string;
  endDate: string;
  limitAmountMinor: number;
  currency: string;
};

export default function BudgetsPageClient() {
  const toast = useApiToasts();
  const [page, setPage] = React.useState(1);
  const [limit] = React.useState(20);

  const listQ = useBudgetsList({ page, limit });
  const createM = useCreateBudget();
  const updateM = useUpdateBudget();
  const deleteM = useDeleteBudget();

  const [edit, setEdit] = React.useState<Budget | null>(null);
  const [usageId, setUsageId] = React.useState<string | null>(null);

  const usageQ = useBudgetUsage(usageId ?? "");
  const showUsage = usageId ? usageQ : null;

  const form = useForm<FormValues>({
    resolver: zodResolver(budgetUpsertSchema) as Resolver<FormValues, any, FormValues>,
    defaultValues: {
      name: "",
      periodType: "monthly",
      categoryId: null,
      startDate: new Date().toISOString(),
      endDate: new Date(Date.now() + 30 * 24 * 3600 * 1000).toISOString(),
      limitAmountMinor: 0,
      currency: "USD",
    },
  });

  const openCreate = () => {
    setEdit(null);
    form.reset({
      name: "",
      periodType: "monthly",
      categoryId: null,
      startDate: new Date().toISOString(),
      endDate: new Date(Date.now() + 30 * 24 * 3600 * 1000).toISOString(),
      limitAmountMinor: 0,
      currency: "USD",
    });
  };

  const openEdit = (b: Budget) => {
    setEdit(b);
    form.reset({
      name: b.name,
      periodType: b.periodType,
      categoryId: b.categoryId ?? null,
      startDate: b.startDate,
      endDate: b.endDate,
      limitAmountMinor: b.limitAmountMinor,
      currency: b.currency,
    });
  };

  const submit = async (values: FormValues) => {
    try {
      if (!edit) {
        await createM.mutateAsync(values);
        toast.showSuccess("Budget created");
      } else {
        await updateM.mutateAsync({
          id: edit.id,
          version: edit.version,
          patch: values,
        });
        toast.showSuccess("Budget updated");
      }
    } catch (e) {
      toast.showError(e, "Budget error");
    }
  };

  const cols: ColumnDef<Budget>[] = [
    { key: "name", header: "Budget", cell: (b) => b.name },
    { key: "period", header: "Period", cell: (b) => b.periodType },
    {
      key: "window",
      header: "Window",
      cell: (b) => (
        <div className="text-xs text-muted-foreground">
          {new Date(b.startDate).toLocaleDateString()} →{" "}
          {new Date(b.endDate).toLocaleDateString()}
        </div>
      ),
    },
    {
      key: "limit",
      header: "Limit",
      cell: (b) =>
        formatMoneyMinor({
          amountMinor: b.limitAmountMinor,
          currency: b.currency,
        }),
    },
    {
      key: "actions",
      header: "",
      className: "w-[1%] whitespace-nowrap",
      cell: (b) => (
        <div className="flex items-center justify-end gap-2">
          <Button variant="outline" size="sm" onClick={() => setUsageId(b.id)}>
            Usage
          </Button>
          <Button variant="outline" size="sm" onClick={() => openEdit(b)}>
            Edit
          </Button>
          <ConfirmDialog
            title="Delete budget?"
            description="This will soft-delete the budget (version-checked)."
            confirmLabel="Delete"
            variant="destructive"
            onConfirm={async () => {
              try {
                await deleteM.mutateAsync({ id: b.id, version: b.version });
                toast.showSuccess("Budget deleted");
              } catch (e) {
                toast.showError(e, "Delete failed");
              }
            }}
          >
            <Button variant="destructive" size="sm">
              Delete
            </Button>
          </ConfirmDialog>
        </div>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        title="Budgets"
        description="Plan spending, monitor usage, and catch overspends early."
        actions={
          <Dialog>
            <DialogTrigger asChild>
              <Button onClick={openCreate}>New budget</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>{edit ? "Edit budget" : "New budget"}</DialogTitle>
                <DialogDescription>
                  Budget usage is calculated from transactions (minor units).
                </DialogDescription>
              </DialogHeader>
              <form className="grid gap-4" onSubmit={form.handleSubmit(submit)}>
                <div className="grid gap-2">
                  <Label htmlFor="name">Name</Label>
                  <Input id="name" {...form.register("name")} />
                  <FieldError msg={form.formState.errors.name?.message} />
                </div>
                <div className="grid gap-2 md:grid-cols-2">
                  <div className="grid gap-2">
                    <Label htmlFor="periodType">Period</Label>
                    <select
                      id="periodType"
                      className="h-10 rounded-md border bg-background px-3 text-sm"
                      {...form.register("periodType")}
                    >
                      <option value="monthly">monthly</option>
                      <option value="weekly">weekly</option>
                      <option value="custom">custom</option>
                    </select>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="currency">Currency</Label>
                    <Input id="currency" {...form.register("currency")} />
                  </div>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="categoryId">Category ID (optional)</Label>
                  <Input
                    id="categoryId"
                    {...form.register("categoryId")}
                    placeholder="uuid or empty for overall spend"
                  />
                </div>
                <div className="grid gap-2 md:grid-cols-2">
                  <div className="grid gap-2">
                    <Label htmlFor="startDate">Start date (RFC3339)</Label>
                    <Input id="startDate" {...form.register("startDate")} />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="endDate">End date (RFC3339)</Label>
                    <Input id="endDate" {...form.register("endDate")} />
                  </div>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="limitAmountMinor">Limit (minor)</Label>
                  <Input
                    id="limitAmountMinor"
                    type="number"
                    {...form.register("limitAmountMinor", {
                      valueAsNumber: true,
                    })}
                  />
                  <FieldError
                    msg={form.formState.errors.limitAmountMinor?.message}
                  />
                </div>
                <div className="flex items-center justify-end gap-2">
                  <Button
                    type="submit"
                    disabled={createM.isPending || updateM.isPending}
                  >
                    {edit ? "Save" : "Create"}
                  </Button>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        }
      />

      {listQ.isLoading ? <LoadingSkeleton /> : null}
      {listQ.isError ? (
        <EmptyState
          title="Couldn’t load budgets"
          description="Check gateway and auth."
          actionLabel="Retry"
          onAction={() => listQ.refetch()}
        />
      ) : null}
      {listQ.data && listQ.data.items.length === 0 ? (
        <EmptyState
          title="No budgets"
          description="Create a budget to start tracking limits."
        />
      ) : null}
      {listQ.data && listQ.data.items.length > 0 ? (
        <div className="space-y-4">
          <DataTable
            columns={cols}
            rows={listQ.data.items}
            rowKey={(b) => b.id}
          />
          <div className="flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
              Page {listQ.data.page} · {listQ.data.items.length} shown ·{" "}
              {listQ.data.total} total
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                Prev
              </Button>
              <Button
                variant="outline"
                disabled={
                  listQ.data.page * listQ.data.limit >= listQ.data.total
                }
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        </div>
      ) : null}

      {showUsage ? (
        <div className="mt-6 rounded-2xl border bg-card p-6 shadow-soft">
          <div className="flex items-start justify-between gap-4">
            <div>
              <div className="text-base font-semibold">Budget usage</div>
              <div className="mt-1 text-xs text-muted-foreground">
                {usageId}
              </div>
            </div>
            <Button variant="secondary" onClick={() => setUsageId(null)}>
              Close
            </Button>
          </div>

          {showUsage.isLoading ? (
            <div className="mt-4">
              <LoadingSkeleton />
            </div>
          ) : null}
          {showUsage.isError ? (
            <div className="mt-4">
              <EmptyState
                title="Couldn’t load usage"
                actionLabel="Retry"
                onAction={() => showUsage.refetch()}
              />
            </div>
          ) : null}
          {showUsage.data ? (
            <div className="mt-4 grid gap-3 md:grid-cols-3">
              <div className="md:col-span-3">
                <StatusBadge
                  variant={showUsage.data.isOverspent ? "warning" : "success"}
                  label={showUsage.data.isOverspent ? "Overspent" : "On track"}
                />
              </div>
              <Metric
                label="Spent"
                value={formatMoneyMinor({
                  amountMinor: showUsage.data.spentMinor,
                  currency: showUsage.data.currency,
                })}
              />
              <Metric
                label="Remaining"
                value={formatMoneyMinor({
                  amountMinor: showUsage.data.remainingMinor,
                  currency: showUsage.data.currency,
                })}
              />
              <Metric
                label="Overspent"
                value={formatMoneyMinor({
                  amountMinor: showUsage.data.overspentMinor,
                  currency: showUsage.data.currency,
                })}
              />
              <div className="md:col-span-3">
                <div className="text-xs font-semibold text-muted-foreground">
                  Related
                </div>
                <div className="mt-1 text-sm">
                  <Link
                    className="underline underline-offset-4"
                    href={`/transactions?from=${encodeURIComponent(showUsage.data.from)}&to=${encodeURIComponent(showUsage.data.to)}&currency=${encodeURIComponent(showUsage.data.currency)}`}
                  >
                    View transactions in this budget window
                  </Link>
                </div>
              </div>
            </div>
          ) : null}
        </div>
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
