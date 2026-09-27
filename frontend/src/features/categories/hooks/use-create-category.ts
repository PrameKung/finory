"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  categoryQueryKeys,
  createCategory,
} from "@/features/categories/api/categories-api";

export function useCreateCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createCategory,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: categoryQueryKeys.all });
    },
  });
}
