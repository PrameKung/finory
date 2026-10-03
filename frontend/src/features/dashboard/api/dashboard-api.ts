import { categoryDistributionSchema } from "@/features/dashboard/schemas/category-distribution-schema";
import { monthlyComparisonSchema } from "@/features/dashboard/schemas/monthly-comparison-schema";
import { monthlySummarySchema } from "@/features/dashboard/schemas/monthly-summary-schema";
import { trendSeriesSchema } from "@/features/dashboard/schemas/trend-series-schema";
import type { CategoryDistribution } from "@/features/dashboard/types/category-distribution";
import type { MonthlyComparison } from "@/features/dashboard/types/monthly-comparison";
import type { MonthlySummary } from "@/features/dashboard/types/monthly-summary";
import type { TrendSeries } from "@/features/dashboard/types/trend-series";
import { apiRequest } from "@/lib/api/client";

const monthlySummaryPath = "/api/v1/analytics/summary";
const categoryDistributionPath = "/api/v1/analytics/categories";
const trendSeriesPath = "/api/v1/analytics/trends";
const monthlyComparisonPath = "/api/v1/analytics/monthly";

export const dashboardQueryKeys = {
  all: ["dashboard"] as const,
  monthlySummary: () => ["dashboard", "monthly-summary"] as const,
  categoryDistribution: () => ["dashboard", "category-distribution"] as const,
  trendSeries: () => ["dashboard", "trend-series"] as const,
  monthlyComparison: () => ["dashboard", "monthly-comparison"] as const,
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

export async function getTrendSeries(): Promise<TrendSeries> {
  const data = await apiRequest<unknown>(trendSeriesPath, {
    cache: "no-store",
  });

  return trendSeriesSchema.parse(data);
}

export async function getMonthlyComparison(): Promise<MonthlyComparison> {
  const data = await apiRequest<unknown>(monthlyComparisonPath, {
    cache: "no-store",
  });

  return monthlyComparisonSchema.parse(data);
}
