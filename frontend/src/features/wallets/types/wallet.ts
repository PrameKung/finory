import type { z } from "zod";

import type {
  createWalletSchema,
  updateWalletSchema,
  walletSchema,
  walletTypeSchema,
} from "@/features/wallets/schemas/wallet-schema";

export type Wallet = z.infer<typeof walletSchema>;
export type WalletType = z.infer<typeof walletTypeSchema>;
export type CreateWalletInput = z.infer<typeof createWalletSchema>;
export type UpdateWalletInput = z.infer<typeof updateWalletSchema>;
