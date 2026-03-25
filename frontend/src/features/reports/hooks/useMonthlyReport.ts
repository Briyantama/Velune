"use client";

import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/src/lib/query/keys";
import { reportsApi } from "@/src/features/reports/services/reportsApi";

export function useMonthlyReport(args: { year: number; month: number; currency: string }) {
  return useQuery({
    queryKey: queryKeys.monthlyReport(args),
    queryFn: () => reportsApi.monthly(args)
  });
}

