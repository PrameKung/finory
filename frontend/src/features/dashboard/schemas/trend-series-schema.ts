import { z } from "zod";

const amountSchema = z.string().regex(/^\d+(\.\d{1,4})?$/);

const trendPointSchema = z.object({
  date: z.iso.date(),
  income: amountSchema,
  expense: amountSchema,
});

export const trendSeriesSchema = z.object({
  month: z.string().regex(/^\d{4}-(0[1-9]|1[0-2])$/),
  granularity: z.literal("day"),
  points: z.array(trendPointSchema),
});
