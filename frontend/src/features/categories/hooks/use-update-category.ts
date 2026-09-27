"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  categoryQueryKeys,
  updateCategory,
} from "@/features/categories/api/categories-api";
import type { UpdateCategoryInput } from "@/features/categories/types/category";

type UpdateCategoryVariables = {
  id: string;
  input: UpdateCategoryInput;
};

export function useUpdateCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, input }: UpdateCategoryVariables) =>
      updateCategory(id, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: categoryQueryKeys.all });
    },
  });
}
