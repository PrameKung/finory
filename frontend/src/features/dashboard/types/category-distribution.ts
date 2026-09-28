import type { z } from "zod";

import type { categoryDistributionSchema } from "@/features/dashboard/schemas/category-distribution-schema";

export type CategoryDistribution = z.infer<typeof categoryDistributionSchema>;
