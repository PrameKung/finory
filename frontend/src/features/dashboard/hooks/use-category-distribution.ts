"use client";

import { useQuery } from "@tanstack/react-query";

import {
  dashboardQueryKeys,
  getCategoryDistribution,
} from "@/features/dashboard/api/dashboard-api";

export function useCategoryDistribution() {
  return useQuery({
    queryKey: dashboardQueryKeys.categoryDistribution(),
    queryFn: getCategoryDistribution,
  });
}
