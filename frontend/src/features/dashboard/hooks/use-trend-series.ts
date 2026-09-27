"use client";

import { useQuery } from "@tanstack/react-query";

import {
  dashboardQueryKeys,
  getTrendSeries,
} from "@/features/dashboard/api/dashboard-api";

export function useTrendSeries() {
  return useQuery({
    queryKey: dashboardQueryKeys.trendSeries(),
    queryFn: getTrendSeries,
  });
}
