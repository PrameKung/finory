"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import {
  transactionQueryKeys,
  updateTransaction,
} from "@/features/transactions/api/transactions-api";
import type { UpdateTransactionInput } from "@/features/transactions/types/transaction";

type UpdateTransactionVariables = {
  id: string;
  input: UpdateTransactionInput;
};

export function useUpdateTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, input }: UpdateTransactionVariables) =>
      updateTransaction(id, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: transactionQueryKeys.all,
      });
      toast.success("Transaction updated");
    },
  });
}
