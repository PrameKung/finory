import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { CreateWalletSheet } from "@/features/wallets/components/create-wallet-sheet";
import { WalletList } from "@/features/wallets/components/wallet-list";

export const metadata: Metadata = {
  title: "Wallets",
};

export default function WalletsPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        title="Wallets"
        description="Manage the accounts where you keep your money."
        action={<CreateWalletSheet />}
      />

      <WalletList />
    </div>
  );
}
