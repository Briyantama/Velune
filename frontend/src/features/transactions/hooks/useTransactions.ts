"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/src/lib/query/keys";
import { transactionsApi, type ListTransactionsArgs } from "@/src/features/transactions/services/transactionsApi";

export function useTransactionsList(args: ListTransactionsArgs) {
  return useQuery({
    queryKey: queryKeys.transactions(args),
    queryFn: () => transactionsApi.list(args)
  });
}

export function useCreateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: transactionsApi.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions"] });
    }
  });
}

export function useUpdateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: transactionsApi.update,
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: queryKeys.transaction(vars.id) });
    }
  });
}

export function useDeleteTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: transactionsApi.delete,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions"] });
    }
  });
}

