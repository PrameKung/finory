import type { Metadata } from "next";

import { CreateWalletSheet } from "@/features/wallets/components/create-wallet-sheet";
import { WalletList } from "@/features/wallets/components/wallet-list";

export const metadata: Metadata = {
  title: "Wallets",
};

export default function WalletsPage() {
  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            Wallets
          </h1>
          <p className="text-sm text-muted-foreground">
            Manage the accounts where you keep your money.
          </p>
        </div>
        <CreateWalletSheet />
      </header>

      <WalletList />
    </div>
  );
}
