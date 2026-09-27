import { walletsSchema } from "@/features/wallets/schemas/wallet-schema";
import type { Wallet } from "@/features/wallets/types/wallet";
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
