"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/src/lib/query/keys";
import { budgetsApi } from "@/src/features/budgets/services/budgetsApi";

export function useBudgetsList(args: { page: number; limit: number; activeOn?: string }) {
  return useQuery({
    queryKey: queryKeys.budgets(args),
    queryFn: () => budgetsApi.list(args)
  });
}

export function useBudgetUsage(id: string) {
  return useQuery({
    queryKey: queryKeys.budgetUsage(id),
    queryFn: () => budgetsApi.usage(id)
  });
}

export function useCreateBudget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: budgetsApi.create,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["budgets"] })
  });
}

export function useUpdateBudget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: budgetsApi.update,
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ["budgets"] });
      qc.invalidateQueries({ queryKey: queryKeys.budgetUsage(vars.id) });
    }
  });
}

export function useDeleteBudget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: budgetsApi.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["budgets"] })
  });
}

