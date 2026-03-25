import { apiGet } from "@/src/lib/api/client";

export type MonthlyCategoryBreakdown = {
  categoryId?: string | null;
  categoryName: string;
  totalMinor: number;
  currency: string;
};

export type MonthlyReport = {
  userId: string;
  year: number;
  month: number;
  incomeMinor: number;
  expenseMinor: number;
  currency: string;
  byCategory: MonthlyCategoryBreakdown[];
  generatedAt: string;
};

export const reportsApi = {
  monthly(args: { year: number; month: number; currency: string }) {
    const p = new URLSearchParams({ year: String(args.year), month: String(args.month), currency: args.currency });
    return apiGet<MonthlyReport>(`/reports/monthly?${p.toString()}`);
  }
};

