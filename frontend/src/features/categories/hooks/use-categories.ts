"use client";

import { useQuery } from "@tanstack/react-query";

import {
  categoryQueryKeys,
  getCategories,
} from "@/features/categories/api/categories-api";

export function useCategories() {
  return useQuery({
    queryKey: categoryQueryKeys.list,
    queryFn: getCategories,
  });
}
