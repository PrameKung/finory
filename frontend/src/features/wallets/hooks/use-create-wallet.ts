"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import {
  createWallet,
  walletQueryKeys,
} from "@/features/wallets/api/wallets-api";

export function useCreateWallet() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createWallet,
    onSuccess: async (wallet) => {
      await queryClient.invalidateQueries({ queryKey: walletQueryKeys.all });
      toast.success(`${wallet.name} created`);
    },
  });
}
