import { categoryDistributionSchema } from "@/features/dashboard/schemas/category-distribution-schema";
import { monthlySummarySchema } from "@/features/dashboard/schemas/monthly-summary-schema";
import type { CategoryDistribution } from "@/features/dashboard/types/category-distribution";
import type { MonthlySummary } from "@/features/dashboard/types/monthly-summary";
import { apiRequest } from "@/lib/api/client";

const monthlySummaryPath = "/api/v1/analytics/summary";
const categoryDistributionPath = "/api/v1/analytics/categories";

export const dashboardQueryKeys = {
  all: ["dashboard"] as const,
  monthlySummary: () => ["dashboard", "monthly-summary"] as const,
  categoryDistribution: () => ["dashboard", "category-distribution"] as const,
};

export async function getMonthlySummary(): Promise<MonthlySummary> {
  const data = await apiRequest<unknown>(monthlySummaryPath, {
    cache: "no-store",
  });

  return monthlySummarySchema.parse(data);
}

export async function getCategoryDistribution(): Promise<CategoryDistribution> {
  const data = await apiRequest<unknown>(categoryDistributionPath, {
    cache: "no-store",
  });

  return categoryDistributionSchema.parse(data);
}
