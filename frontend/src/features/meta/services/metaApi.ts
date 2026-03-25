import { apiGet, apiPost, apiPut, apiDelete } from "@/src/lib/api/client";

export type Account = {
  id: string;
  name: string;
  type: string;
  currency: string;
  balanceMinor: number;
  version: number;
};

export type Category = {
  id: string;
  name: string;
  parentId?: string | null;
  version: number;
};

export type Paged<T> = { items: T[]; total: number; page: number; limit: number };

export const metaApi = {
  listAccounts(args: { page: number; limit: number }) {
    const p = new URLSearchParams({ page: String(args.page), limit: String(args.limit) });
    return apiGet<Paged<Account>>(`/accounts?${p.toString()}`);
  },
  listCategories(args: { page: number; limit: number }) {
    const p = new URLSearchParams({ page: String(args.page), limit: String(args.limit) });
    return apiGet<Paged<Category>>(`/categories?${p.toString()}`);
  },

  createAccount(input: { name: string; type: string; currency: string }) {
    return apiPost<Account>("/accounts", input);
  },
  updateAccount(input: { id: string; version: number; name: string; type: string }) {
    const p = new URLSearchParams({ id: input.id, version: String(input.version) });
    return apiPut<Account>(`/accounts/${input.id}?${p.toString()}`, { name: input.name, type: input.type });
  },
  deleteAccount(input: { id: string; version: number }) {
    const p = new URLSearchParams({ id: input.id, version: String(input.version) });
    return apiDelete<void>(`/accounts/${input.id}?${p.toString()}`);
  }
};

