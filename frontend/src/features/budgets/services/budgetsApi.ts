import { apiDelete, apiGet, apiPost, apiPut } from "@/src/lib/api/client";

export type BudgetPeriodType = "monthly" | "weekly" | "custom";

export type Budget = {
  id: string;
  userId: string;
  name: string;
  periodType: BudgetPeriodType;
  categoryId?: string | null;
  startDate: string;
  endDate: string;
  limitAmountMinor: number;
  currency: string;
  version: number;
};

export type BudgetUsage = {
  budgetId: string;
  from: string;
  to: string;
  currency: string;
  limitAmountMinor: number;
  spentMinor: number;
  remainingMinor: number;
  overspentMinor: number;
  isOverspent: boolean;
};

export type Paged<T> = { items: T[]; total: number; page: number; limit: number };

export const budgetsApi = {
  list(args: { page: number; limit: number; activeOn?: string }) {
    const p = new URLSearchParams({ page: String(args.page), limit: String(args.limit) });
    if (args.activeOn) p.set("activeOn", args.activeOn);
    return apiGet<Paged<Budget>>(`/budgets?${p.toString()}`);
  },

  usage(id: string) {
    const p = new URLSearchParams({ id });
    return apiGet<BudgetUsage>(`/budgets/${id}/usage?${p.toString()}`);
  },

  create(input: {
    name: string;
    periodType: BudgetPeriodType;
    categoryId?: string | null;
    startDate: string;
    endDate: string;
    limitAmountMinor: number;
    currency: string;
  }) {
    return apiPost<Budget>("/budgets", input);
  },

  update(input: {
    id: string;
    version: number;
    patch: {
      name: string;
      periodType: BudgetPeriodType;
      categoryId?: string | null;
      startDate: string;
      endDate: string;
      limitAmountMinor: number;
      currency: string;
    };
  }) {
    const p = new URLSearchParams({ id: input.id, version: String(input.version) });
    return apiPut<Budget>(`/budgets/${input.id}?${p.toString()}`, input.patch);
  },

  delete(input: { id: string; version: number }) {
    const p = new URLSearchParams({ id: input.id, version: String(input.version) });
    return apiDelete<void>(`/budgets/${input.id}?${p.toString()}`);
  }
};

