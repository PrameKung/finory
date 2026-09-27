import {
  walletSchema,
  walletsSchema,
} from "@/features/wallets/schemas/wallet-schema";
import type {
  CreateWalletInput,
  UpdateWalletInput,
  Wallet,
} from "@/features/wallets/types/wallet";
import { apiRequest } from "@/lib/api/client";

const walletsPath = "/api/v1/wallets";

export const walletQueryKeys = {
  all: ["wallets"] as const,
  list: ["wallets", "list"] as const,
};

export async function getWallets(): Promise<Wallet[]> {
  const data = await apiRequest<unknown>(walletsPath, {
    cache: "no-store",
  });

  return walletsSchema.parse(data);
}

export async function createWallet(input: CreateWalletInput): Promise<Wallet> {
  const data = await apiRequest<unknown>(walletsPath, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });

  return walletSchema.parse(data);
}

export async function updateWallet(
  id: string,
  input: UpdateWalletInput,
): Promise<Wallet> {
  const data = await apiRequest<unknown>(`${walletsPath}/${id}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });

  return walletSchema.parse(data);
}

export function deleteWallet(id: string) {
  return apiRequest<void>(`${walletsPath}/${id}`, {
    method: "DELETE",
  });
}
