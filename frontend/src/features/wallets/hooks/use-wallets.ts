"use client";

import { useQuery } from "@tanstack/react-query";

import {
  getWallets,
  walletQueryKeys,
} from "@/features/wallets/api/wallets-api";

export function useWallets() {
  return useQuery({
    queryKey: walletQueryKeys.list,
    queryFn: getWallets,
  });
}
