"use client";

import { useQuery } from "@tanstack/react-query";

import {
  dashboardQueryKeys,
  getMonthlyComparison,
} from "@/features/dashboard/api/dashboard-api";

export function useMonthlyComparison() {
  return useQuery({
    queryKey: dashboardQueryKeys.monthlyComparison(),
    queryFn: getMonthlyComparison,
  });
}
