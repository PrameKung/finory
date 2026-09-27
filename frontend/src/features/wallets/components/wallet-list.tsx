"use client";

import {
  BanknoteIcon,
  LandmarkIcon,
  SmartphoneIcon,
  WalletCardsIcon,
  type LucideIcon,
} from "lucide-react";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { DeleteWalletDialog } from "@/features/wallets/components/delete-wallet-dialog";
import { EditWalletSheet } from "@/features/wallets/components/edit-wallet-sheet";
import { useWallets } from "@/features/wallets/hooks/use-wallets";
import type {
  Wallet,
  WalletType,
} from "@/features/wallets/types/wallet";
import { getApiErrorMessage } from "@/lib/api/client";

const walletTypeDetails: Record<
  WalletType,
  { label: string; icon: LucideIcon }
> = {
  cash: { label: "Cash", icon: BanknoteIcon },
  bank: { label: "Bank account", icon: LandmarkIcon },
  e_wallet: { label: "E-wallet", icon: SmartphoneIcon },
  other: { label: "Other", icon: WalletCardsIcon },
};

function formatBalance(balance: string, currencyCode: string) {
  const numericBalance = Number(balance);

  if (!Number.isFinite(numericBalance)) {
    return `${balance} ${currencyCode}`;
  }

  try {
    return new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: currencyCode,
    }).format(numericBalance);
  } catch {
    return `${balance} ${currencyCode}`;
  }
}

function WalletItem({ wallet }: { wallet: Wallet }) {
  const details = walletTypeDetails[wallet.type];
  const Icon = details.icon;

  return (
    <li className="rounded-xl border bg-card p-5 shadow-sm">
      <div className="flex items-start gap-3">
        <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground">
          <Icon className="size-5" aria-hidden="true" />
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 items-center gap-2">
            <h2 className="truncate text-sm font-medium">{wallet.name}</h2>
            {wallet.isDefault ? (
              <span className="shrink-0 rounded-full bg-muted px-2 py-0.5 text-[0.6875rem] font-medium text-muted-foreground">
                Default
              </span>
            ) : null}
          </div>
          <p className="text-xs text-muted-foreground">{details.label}</p>
        </div>
        {!wallet.isDefault ? (
          <div className="flex shrink-0 items-center gap-1">
            <EditWalletSheet wallet={wallet} />
            <DeleteWalletDialog wallet={wallet} />
          </div>
        ) : null}
      </div>
      <div className="mt-6">
        <p className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
          Balance
        </p>
        <p className="mt-1 font-heading text-2xl font-semibold tracking-tight tabular-nums">
          {formatBalance(wallet.balance, wallet.currencyCode)}
        </p>
      </div>
    </li>
  );
}

export function WalletList() {
  const walletsQuery = useWallets();

  if (walletsQuery.isPending) {
    return (
      <LoadingState
        title="Loading wallets"
        description="Fetching your wallets and current balances."
      />
    );
  }

  if (walletsQuery.isError) {
    return (
      <ErrorState
        title="Could not load wallets"
        description={getApiErrorMessage(walletsQuery.error, {
          defaultMessage:
            "Your wallets are temporarily unavailable. Please try again.",
        })}
        onRetry={() => void walletsQuery.refetch()}
      />
    );
  }

  if (walletsQuery.data.length === 0) {
    return (
      <EmptyState
        title="No wallets yet"
        description="Create a wallet to start tracking where you keep your money."
        icon={<WalletCardsIcon className="size-5" aria-hidden="true" />}
      />
    );
  }

  return (
    <section aria-labelledby="wallet-list-heading">
      <div className="mb-3 flex items-end justify-between gap-4">
        <div>
          <h2
            id="wallet-list-heading"
            className="font-heading text-lg font-semibold tracking-tight"
          >
            Your wallets
          </h2>
          <p className="text-sm text-muted-foreground">
            Balances across all your accounts.
          </p>
        </div>
        <span className="shrink-0 text-sm tabular-nums text-muted-foreground">
          {walletsQuery.data.length}{" "}
          {walletsQuery.data.length === 1 ? "wallet" : "wallets"}
        </span>
      </div>

      <ul className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {walletsQuery.data.map((wallet) => (
          <WalletItem key={wallet.id} wallet={wallet} />
        ))}
      </ul>
    </section>
  );
}
