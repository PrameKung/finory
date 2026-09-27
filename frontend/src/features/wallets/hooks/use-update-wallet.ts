"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  updateWallet,
  walletQueryKeys,
} from "@/features/wallets/api/wallets-api";
import type { UpdateWalletInput } from "@/features/wallets/types/wallet";

type UpdateWalletVariables = {
  id: string;
  input: UpdateWalletInput;
};

export function useUpdateWallet() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, input }: UpdateWalletVariables) =>
      updateWallet(id, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: walletQueryKeys.all });
    },
  });
}
