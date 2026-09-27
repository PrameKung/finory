import { z } from "zod";

const amountSchema = z.string().regex(/^\d+(\.\d{1,4})?$/);

export const categoryRankSchema = z.object({
  rank: z.number().int().positive(),
  categoryId: z.uuid(),
  name: z.string().min(1),
  icon: z.string().nullable(),
  color: z.string().nullable(),
  amount: amountSchema,
  percentage: z.string().regex(/^\d+(\.\d{1,2})?$/),
});

export const categoryDistributionSchema = z.object({
  month: z.string().regex(/^\d{4}-(0[1-9]|1[0-2])$/),
  totalExpense: amountSchema,
  categories: z.array(categoryRankSchema),
});
