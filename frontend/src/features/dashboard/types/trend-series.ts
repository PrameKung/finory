import type { z } from "zod";

import type {
  trendPointSchema,
  trendSeriesSchema,
} from "@/features/dashboard/schemas/trend-series-schema";

export type TrendPoint = z.infer<typeof trendPointSchema>;
export type TrendSeries = z.infer<typeof trendSeriesSchema>;
