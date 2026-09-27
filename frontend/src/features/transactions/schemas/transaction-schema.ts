import { z } from "zod";

export const transactionTypeSchema = z.enum(["income", "expense"]);

export const transactionSchema = z.object({
  id: z.uuid(),
  categoryId: z.uuid(),
  walletId: z.uuid(),
  type: transactionTypeSchema,
  amount: z.string(),
  description: z.string().nullable(),
  transactionDate: z.iso.date(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export const transactionsSchema = z.array(transactionSchema);
