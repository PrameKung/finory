"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import {
  categoryQueryKeys,
  createCategory,
} from "@/features/categories/api/categories-api";

export function useCreateCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createCategory,
    onSuccess: async (category) => {
      await queryClient.invalidateQueries({ queryKey: categoryQueryKeys.all });
      toast.success(`${category.name} created`);
    },
  });
}
