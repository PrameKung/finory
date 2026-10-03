import type { z } from "zod";

import type { monthlySummarySchema } from "@/features/dashboard/schemas/monthly-summary-schema";

export type MonthlySummary = z.infer<typeof monthlySummarySchema>;
