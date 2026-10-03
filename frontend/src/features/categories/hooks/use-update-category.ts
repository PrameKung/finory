"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

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
    onSuccess: async (category) => {
      await queryClient.invalidateQueries({ queryKey: categoryQueryKeys.all });
      toast.success(`${category.name} updated`);
    },
  });
}
