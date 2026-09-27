"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import {
  createTransaction,
  transactionQueryKeys,
} from "@/features/transactions/api/transactions-api";

export function useCreateTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createTransaction,
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: transactionQueryKeys.all,
      });
      toast.success("Transaction added");
    },
  });
}
