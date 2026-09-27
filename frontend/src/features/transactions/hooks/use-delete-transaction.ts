"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import {
  deleteTransaction,
  transactionQueryKeys,
} from "@/features/transactions/api/transactions-api";

export function useDeleteTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteTransaction,
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: transactionQueryKeys.all,
      });
      toast.success("Transaction deleted");
    },
  });
}
