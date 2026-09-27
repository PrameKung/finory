import type { z } from "zod";

import type { monthlyComparisonSchema } from "@/features/dashboard/schemas/monthly-comparison-schema";

export type MonthlyComparison = z.infer<typeof monthlyComparisonSchema>;
