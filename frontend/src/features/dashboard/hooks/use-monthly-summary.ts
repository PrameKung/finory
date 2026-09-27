"use client";

import { useQuery } from "@tanstack/react-query";

import {
  dashboardQueryKeys,
  getMonthlySummary,
} from "@/features/dashboard/api/dashboard-api";

export function useMonthlySummary() {
  return useQuery({
    queryKey: dashboardQueryKeys.monthlySummary(),
    queryFn: getMonthlySummary,
  });
}
