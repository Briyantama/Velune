import { z } from "zod";

export const budgetPeriodSchema = z.enum(["monthly", "weekly", "custom"]);

export const budgetUpsertSchema = z.object({
  name: z.string().trim().min(1).max(200),
  periodType: budgetPeriodSchema,
  categoryId: z.string().uuid().nullable().optional(),
  startDate: z.string().min(1),
  endDate: z.string().min(1),
  limitAmountMinor: z.coerce.number().int().min(0),
  currency: z.string().trim().length(3)
});
