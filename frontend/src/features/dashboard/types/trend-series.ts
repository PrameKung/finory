import type { z } from "zod";

import type { trendSeriesSchema } from "@/features/dashboard/schemas/trend-series-schema";

export type TrendSeries = z.infer<typeof trendSeriesSchema>;
