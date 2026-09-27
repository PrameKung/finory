"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  categoryQueryKeys,
  deleteCategory,
} from "@/features/categories/api/categories-api";

export function useDeleteCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteCategory,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: categoryQueryKeys.all });
    },
  });
}
