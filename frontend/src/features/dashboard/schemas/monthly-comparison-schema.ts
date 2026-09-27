import { z } from "zod";

const monthSchema = z.string().regex(/^\d{4}-(0[1-9]|1[0-2])$/);
const amountSchema = z.string().regex(/^\d+(\.\d{1,4})?$/);
const signedAmountSchema = z.string().regex(/^-?\d+(\.\d{1,4})?$/);
const percentageSchema = z.string().regex(/^-?\d+(\.\d{1,2})?$/);

const monthlyTotalsSchema = z.object({
  income: amountSchema,
  expense: amountSchema,
});

const amountChangeSchema = z.object({
  amount: signedAmountSchema,
  percentage: percentageSchema.nullable(),
});

export const monthlyComparisonSchema = z.object({
  month: monthSchema,
  previousMonth: monthSchema,
  current: monthlyTotalsSchema,
  previous: monthlyTotalsSchema,
  changes: z.object({
    income: amountChangeSchema,
    expense: amountChangeSchema,
  }),
});
