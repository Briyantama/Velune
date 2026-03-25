import { z } from "zod";

export const transactionTypeSchema = z.enum(["income", "expense", "transfer", "adjustment"]);

export const transactionCreateSchema = z.object({
  accountId: z.string().uuid(),
  categoryId: z.string().uuid().nullable().optional(),
  counterpartyAccountId: z.string().uuid().nullable().optional(),
  amountMinor: z.coerce.number().int(),
  currency: z.string().trim().length(3),
  type: transactionTypeSchema,
  description: z.string().trim().max(2000).optional().default(""),
  occurredAt: z.string().min(1)
});

