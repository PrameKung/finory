import type { z } from "zod";

import type {
  categoryDistributionSchema,
  categoryRankSchema,
} from "@/features/dashboard/schemas/category-distribution-schema";

export type CategoryDistribution = z.infer<typeof categoryDistributionSchema>;
export type CategoryRank = z.infer<typeof categoryRankSchema>;
