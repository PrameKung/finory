"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  deleteWallet,
  walletQueryKeys,
} from "@/features/wallets/api/wallets-api";

export function useDeleteWallet() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteWallet,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: walletQueryKeys.all });
    },
  });
}
