export const queryKeys = {
  me: () => ["me"] as const,
  accounts: (args: { page: number; limit: number }) => ["accounts", args] as const,
  categories: (args: { page: number; limit: number }) => ["categories", args] as const,
  transactions: (args: Record<string, unknown>) => ["transactions", args] as const,
  transaction: (id: string) => ["transaction", id] as const,
  transactionSummary: (args: { from: string; to: string; currency: string }) => ["transactionSummary", args] as const,
  budgets: (args: { page: number; limit: number; activeOn?: string }) => ["budgets", args] as const,
  budgetUsage: (id: string) => ["budgetUsage", id] as const,
  monthlyReport: (args: { year: number; month: number; currency: string }) => ["monthlyReport", args] as const,
  notificationsPing: () => ["notificationsPing"] as const,
  adminHealth: () => ["adminHealth"] as const,
  adminDlq: (args: { limit: number }) => ["adminDlq", args] as const,
  adminOutbox: (args: { service: string; status?: string; limit: number }) => ["adminOutbox", args] as const,
  adminReconcileLogs: (args: { service: string; type?: string; limit: number }) => ["adminReconcileLogs", args] as const
};
