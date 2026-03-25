"use client";

import * as React from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { PageHeader } from "@/src/components/common/page-header";
import { FilterBar } from "@/src/components/common/filter-bar";
import { DataTable, type ColumnDef } from "@/src/components/common/data-table";
import { EmptyState } from "@/src/components/common/empty-state";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { ConfirmDialog } from "@/src/components/common/confirm-dialog";
import { Button } from "@/src/components/ui/button";
import { Input } from "@/src/components/ui/input";
import { Label } from "@/src/components/ui/label";
import { Textarea } from "@/src/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/src/components/ui/dialog";
import { formatMoneyMinor } from "@/src/lib/money/format";
import { useApiToasts } from "@/src/lib/api/toast";
import {
  useCreateTransaction,
  useDeleteTransaction,
  useTransactionsList,
  useUpdateTransaction,
} from "@/src/features/transactions/hooks/useTransactions";
import { transactionCreateSchema } from "@/src/features/transactions/schema/transactionSchemas";
import type {
  Transaction,
  TransactionType,
} from "@/src/features/transactions/services/transactionsApi";

type FormValues = {
  accountId: string;
  categoryId?: string | null;
  counterpartyAccountId?: string | null;
  amountMinor: number;
  currency: string;
  type: TransactionType;
  description?: string;
  occurredAt: string;
};

export default function TransactionsPageClient() {
  const toast = useApiToasts();
  const sp = useSearchParams();

  const [page, setPage] = React.useState(1);
  const [limit] = React.useState(20);
  const [search, setSearch] = React.useState("");

  const [type, setType] = React.useState<TransactionType | "">("");
  const [currency, setCurrency] = React.useState("USD");
  const [from, setFrom] = React.useState("");
  const [to, setTo] = React.useState("");
  const [accountId, setAccountId] = React.useState("");
  const [categoryId, setCategoryId] = React.useState("");

  React.useEffect(() => {
    // Initialize from URL query (used by budget->transactions deep links).
    const qCurrency = sp.get("currency");
    const qFrom = sp.get("from");
    const qTo = sp.get("to");
    const qAccount = sp.get("accountId");
    const qCategory = sp.get("categoryId");
    if (qCurrency) setCurrency(qCurrency);
    if (qFrom) setFrom(qFrom);
    if (qTo) setTo(qTo);
    if (qAccount) setAccountId(qAccount);
    if (qCategory) setCategoryId(qCategory);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const listQ = useTransactionsList({
    page,
    limit,
    type: type || undefined,
    currency: currency || undefined,
    from: from || undefined,
    to: to || undefined,
    accountId: accountId || undefined,
    categoryId: categoryId || undefined,
  });

  const items = React.useMemo(() => {
    const rows = listQ.data?.items ?? [];
    const q = search.trim().toLowerCase();
    if (!q) return rows;
    return rows.filter((t) => (t.description ?? "").toLowerCase().includes(q));
  }, [listQ.data, search]);

  const createM = useCreateTransaction();
  const updateM = useUpdateTransaction();
  const deleteM = useDeleteTransaction();

  const [edit, setEdit] = React.useState<Transaction | null>(null);

  const form = useForm<FormValues>({
    resolver: zodResolver(transactionCreateSchema),
    defaultValues: {
      accountId: "",
      categoryId: null,
      counterpartyAccountId: null,
      amountMinor: 0,
      currency: "USD",
      type: "expense",
      description: "",
      occurredAt: new Date().toISOString(),
    },
  });

  const openCreate = () => {
    setEdit(null);
    form.reset({
      accountId: "",
      categoryId: null,
      counterpartyAccountId: null,
      amountMinor: 0,
      currency,
      type: "expense",
      description: "",
      occurredAt: new Date().toISOString(),
    });
  };

  const openEdit = (t: Transaction) => {
    setEdit(t);
    form.reset({
      accountId: t.accountId,
      categoryId: t.categoryId ?? null,
      counterpartyAccountId: t.counterpartyAccountId ?? null,
      amountMinor: t.amountMinor,
      currency: t.currency,
      type: t.type,
      description: t.description ?? "",
      occurredAt: t.occurredAt,
    });
  };

  const submit = async (values: FormValues) => {
    try {
      if (!edit) {
        await createM.mutateAsync(values);
        toast.showSuccess("Transaction created");
      } else {
        await updateM.mutateAsync({
          id: edit.id,
          version: edit.version,
          patch: values,
        });
        toast.showSuccess("Transaction updated");
      }
    } catch (e) {
      toast.showError(e, "Transaction error");
    }
  };

  const cols: ColumnDef<Transaction>[] = [
    {
      key: "occurredAt",
      header: "Date",
      cell: (t) => new Date(t.occurredAt).toLocaleString(),
    },
    {
      key: "type",
      header: "Type",
      cell: (t) => t.type,
    },
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
        <div className="max-w-[420px] truncate">
          <Link
            className="underline underline-offset-4"
            href={`/transactions/${t.id}`}
          >
            {t.description || "—"}
          </Link>
        </div>
      ),
    },
    {
      key: "actions",
      header: "",
      className: "w-[1%] whitespace-nowrap",
      cell: (t) => (
        <div className="flex items-center justify-end gap-2">
          <Button variant="outline" size="sm" onClick={() => openEdit(t)}>
            Edit
          </Button>
          <ConfirmDialog
            title="Delete transaction?"
            description="This will soft-delete the transaction (version-checked)."
            confirmLabel="Delete"
            variant="destructive"
            onConfirm={async () => {
              try {
                await deleteM.mutateAsync({ id: t.id, version: t.version });
                toast.showSuccess("Transaction deleted");
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
        title="Transactions"
        description="Search, filter, and maintain your ledger."
        actions={
          <Dialog>
            <DialogTrigger asChild>
              <Button onClick={openCreate}>New transaction</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>
                  {edit ? "Edit transaction" : "New transaction"}
                </DialogTitle>
                <DialogDescription>
                  Amounts are stored in minor units (integers).
                </DialogDescription>
              </DialogHeader>

              <form className="grid gap-4" onSubmit={form.handleSubmit(submit)}>
                <div className="grid gap-2">
                  <Label htmlFor="accountId">Account ID</Label>
                  <Input
                    id="accountId"
                    {...form.register("accountId")}
                    placeholder="uuid"
                  />
                  <FieldError msg={form.formState.errors.accountId?.message} />
                </div>

                <div className="grid gap-2 md:grid-cols-2">
                  <div className="grid gap-2">
                    <Label htmlFor="type">Type</Label>
                    <select
                      id="type"
                      className="h-10 rounded-md border bg-background px-3 text-sm"
                      {...form.register("type")}
                    >
                      <option value="expense">expense</option>
                      <option value="income">income</option>
                      <option value="transfer">transfer</option>
                      <option value="adjustment">adjustment</option>
                    </select>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="currency">Currency</Label>
                    <Input
                      id="currency"
                      {...form.register("currency")}
                      placeholder="USD"
                    />
                  </div>
                </div>

                <div className="grid gap-2 md:grid-cols-2">
                  <div className="grid gap-2">
                    <Label htmlFor="amountMinor">Amount (minor)</Label>
                    <Input
                      id="amountMinor"
                      type="number"
                      {...form.register("amountMinor", { valueAsNumber: true })}
                    />
                    <FieldError
                      msg={form.formState.errors.amountMinor?.message}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="occurredAt">Occurred at (RFC3339)</Label>
                    <Input id="occurredAt" {...form.register("occurredAt")} />
                    <FieldError
                      msg={form.formState.errors.occurredAt?.message}
                    />
                  </div>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="categoryId">Category ID (optional)</Label>
                  <Input
                    id="categoryId"
                    {...form.register("categoryId")}
                    placeholder="uuid or empty"
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="counterpartyAccountId">
                    Counterparty Account ID (transfer only)
                  </Label>
                  <Input
                    id="counterpartyAccountId"
                    {...form.register("counterpartyAccountId")}
                    placeholder="uuid or empty"
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    {...form.register("description")}
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

      <FilterBar>
        <div className="grid gap-2">
          <Label>Search (this page)</Label>
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="description contains…"
          />
        </div>
        <div className="grid gap-2">
          <Label>Type</Label>
          <select
            className="h-10 rounded-md border bg-background px-3 text-sm"
            value={type}
            onChange={(e) => setType(e.target.value as any)}
          >
            <option value="">all</option>
            <option value="expense">expense</option>
            <option value="income">income</option>
            <option value="transfer">transfer</option>
            <option value="adjustment">adjustment</option>
          </select>
        </div>
        <div className="grid gap-2">
          <Label>Currency</Label>
          <Input
            value={currency}
            onChange={(e) => setCurrency(e.target.value)}
            placeholder="USD"
          />
        </div>
        <div className="grid gap-2">
          <Label>From (RFC3339)</Label>
          <Input
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            placeholder="2026-01-01T00:00:00Z"
          />
        </div>
        <div className="grid gap-2">
          <Label>To (RFC3339)</Label>
          <Input
            value={to}
            onChange={(e) => setTo(e.target.value)}
            placeholder="2026-02-01T00:00:00Z"
          />
        </div>
        <div className="grid gap-2">
          <Label>Account ID</Label>
          <Input
            value={accountId}
            onChange={(e) => setAccountId(e.target.value)}
            placeholder="uuid"
          />
        </div>
        <div className="grid gap-2">
          <Label>Category ID</Label>
          <Input
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
            placeholder="uuid"
          />
        </div>
      </FilterBar>

      {listQ.isLoading ? <LoadingSkeleton /> : null}
      {listQ.isError ? (
        <EmptyState
          title="Couldn’t load transactions"
          description="Check gateway base URL, auth, and service health."
          actionLabel="Retry"
          onAction={() => listQ.refetch()}
        />
      ) : null}
      {listQ.data && items.length === 0 ? (
        <EmptyState
          title="No transactions"
          description="Create your first transaction to begin tracking."
        />
      ) : null}
      {listQ.data && items.length > 0 ? (
        <div className="space-y-4">
          <DataTable columns={cols} rows={items} rowKey={(t) => t.id} />
          <div className="flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
              Page {listQ.data.page} · {items.length} shown · {listQ.data.total}{" "}
              total
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
    </div>
  );
}

function FieldError({ msg }: { msg?: string }) {
  if (!msg) return null;
  return <div className="text-xs text-destructive">{msg}</div>;
}
