import type { z } from "zod";

import type {
  walletSchema,
  walletTypeSchema,
} from "@/features/wallets/schemas/wallet-schema";

export type Wallet = z.infer<typeof walletSchema>;
export type WalletType = z.infer<typeof walletTypeSchema>;
