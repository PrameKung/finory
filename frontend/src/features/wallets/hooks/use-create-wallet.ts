"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  createWallet,
  walletQueryKeys,
} from "@/features/wallets/api/wallets-api";

export function useCreateWallet() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createWallet,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: walletQueryKeys.all });
    },
  });
}
