import { z } from "zod";

const amountSchema = z.string().regex(/^\d+(\.\d{1,4})?$/);
const signedAmountSchema = z.string().regex(/^-?\d+(\.\d{1,4})?$/);

export const monthlySummarySchema = z.object({
  month: z.string().regex(/^\d{4}-(0[1-9]|1[0-2])$/),
  income: amountSchema,
  expense: amountSchema,
  netBalance: signedAmountSchema,
});
