import { apiDelete, apiGet, apiPatch, apiPost } from "@/src/lib/api/client";

export type TransactionType = "income" | "expense" | "transfer" | "adjustment";

export type Transaction = {
  id: string;
  userId: string;
  accountId: string;
  categoryId?: string | null;
  counterpartyAccountId?: string | null;
  amountMinor: number;
  currency: string;
  type: TransactionType;
  description: string;
  occurredAt: string;
  version: number;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
};

export type Paged<T> = {
  items: T[];
  total: number;
  page: number;
  limit: number;
};

export type ListTransactionsArgs = {
  page: number;
  limit: number;
  accountId?: string;
  categoryId?: string;
  type?: TransactionType;
  from?: string;
  to?: string;
  currency?: string;
};

export type TransactionCreateInput = {
  accountId: string;
  categoryId?: string | null;
  counterpartyAccountId?: string | null;
  amountMinor: number;
  currency: string;
  type: TransactionType;
  description?: string;
  occurredAt: string;
};

export type TransactionUpdateInput = TransactionCreateInput;

export const transactionsApi = {
  list(args: ListTransactionsArgs) {
    const p = new URLSearchParams();
    p.set("page", String(args.page));
    p.set("limit", String(args.limit));
    if (args.accountId) p.set("accountId", args.accountId);
    if (args.categoryId) p.set("categoryId", args.categoryId);
    if (args.type) p.set("type", args.type);
    if (args.from) p.set("from", args.from);
    if (args.to) p.set("to", args.to);
    if (args.currency) p.set("currency", args.currency);
    return apiGet<Paged<Transaction>>(`/transactions?${p.toString()}`);
  },

  get(id: string) {
    const p = new URLSearchParams();
    p.set("id", id);
    return apiGet<Transaction>(`/transactions/${id}?${p.toString()}`);
  },

  create(input: TransactionCreateInput) {
    return apiPost<Transaction>("/transactions", input);
  },

  update(input: { id: string; version: number; patch: TransactionUpdateInput }) {
    const p = new URLSearchParams();
    p.set("id", input.id);
    p.set("version", String(input.version));
    return apiPatch<Transaction>(`/transactions/${input.id}?${p.toString()}`, input.patch);
  },

  delete(input: { id: string; version: number }) {
    const p = new URLSearchParams();
    p.set("id", input.id);
    p.set("version", String(input.version));
    return apiDelete<void>(`/transactions/${input.id}?${p.toString()}`);
  }
  ,
  summary(args: { from: string; to: string; currency: string }) {
    const p = new URLSearchParams({ from: args.from, to: args.to, currency: args.currency });
    return apiGet<{ from: string; to: string; currency: string; incomeMinor: number; expenseMinor: number; netMinor: number }>(
      `/transactions/summary?${p.toString()}`
    );
  }
};

