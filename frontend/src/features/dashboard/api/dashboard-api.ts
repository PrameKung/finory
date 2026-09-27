import { monthlySummarySchema } from "@/features/dashboard/schemas/monthly-summary-schema";
import type { MonthlySummary } from "@/features/dashboard/types/monthly-summary";
import { apiRequest } from "@/lib/api/client";

const monthlySummaryPath = "/api/v1/analytics/summary";

export const dashboardQueryKeys = {
  all: ["dashboard"] as const,
  monthlySummary: () => ["dashboard", "monthly-summary"] as const,
};

export async function getMonthlySummary(): Promise<MonthlySummary> {
  const data = await apiRequest<unknown>(monthlySummaryPath, {
    cache: "no-store",
  });

  return monthlySummarySchema.parse(data);
}
