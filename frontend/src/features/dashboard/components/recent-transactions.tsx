"use client";

import { format, parseISO } from "date-fns";
import {
  ArrowDownLeftIcon,
  ArrowLeftRightIcon,
  ArrowRightIcon,
  ArrowUpRightIcon,
} from "lucide-react";
import Link from "next/link";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { Button } from "@/components/ui/button";
import { useCategories } from "@/features/categories/hooks/use-categories";
import { useTransactions } from "@/features/transactions/hooks/use-transactions";
import { formatTransactionAmount } from "@/features/transactions/lib/format-transaction-amount";
import { useWallets } from "@/features/wallets/hooks/use-wallets";
import { cn } from "@/lib/utils";

const recentTransactionLimit = 5;

export function RecentTransactions() {
  const transactionsQuery = useTransactions({});
  const categoriesQuery = useCategories();
  const walletsQuery = useWallets();
  const isPending =
    transactionsQuery.isPending ||
    categoriesQuery.isPending ||
    walletsQuery.isPending;
  const isError =
    transactionsQuery.isError ||
    categoriesQuery.isError ||
    walletsQuery.isError;

  function retry() {
    void Promise.all([
      transactionsQuery.refetch(),
      categoriesQuery.refetch(),
      walletsQuery.refetch(),
    ]);
  }

  if (isPending) {
    return (
      <LoadingState
        title="Loading recent transactions"
        description="Fetching your latest financial activity."
        className="min-h-80"
      />
    );
  }

  if (isError) {
    return (
      <ErrorState
        title="Recent transactions unavailable"
        description="We could not load your latest financial activity."
        className="min-h-80"
        onRetry={retry}
      />
    );
  }

  if (transactionsQuery.data.length === 0) {
    return (
      <EmptyState
        title="No transactions yet"
        description="Your latest income and expenses will appear here."
        icon={<ArrowLeftRightIcon className="size-5" aria-hidden="true" />}
        action={
          <Button asChild variant="outline">
            <Link href="/transactions">Go to transactions</Link>
          </Button>
        }
        className="min-h-80"
      />
    );
  }

  const categoriesById = new Map(
    categoriesQuery.data.map((category) => [category.id, category]),
  );
  const walletsById = new Map(
    walletsQuery.data.map((wallet) => [wallet.id, wallet]),
  );
  const recentTransactions = transactionsQuery.data.slice(
    0,
    recentTransactionLimit,
  );

  return (
    <article className="overflow-hidden rounded-xl border bg-card shadow-sm">
      <header className="flex items-start justify-between gap-4 border-b px-5 py-4">
        <div className="space-y-1">
          <h2 className="font-heading text-lg font-semibold tracking-tight">
            Recent transactions
          </h2>
          <p className="text-sm text-muted-foreground">
            Your latest income and expenses.
          </p>
        </div>
        <Button asChild variant="ghost" size="sm">
          <Link href="/transactions">
            View all
            <ArrowRightIcon aria-hidden="true" />
          </Link>
        </Button>
      </header>

      <ul className="divide-y">
        {recentTransactions.map((transaction) => {
          const category = categoriesById.get(transaction.categoryId);
          const wallet = walletsById.get(transaction.walletId);
          const isIncome = transaction.type === "income";
          const Icon = isIncome ? ArrowDownLeftIcon : ArrowUpRightIcon;

          return (
            <li
              key={transaction.id}
              className="flex items-center gap-3 px-5 py-4"
            >
              <span
                className={cn(
                  "flex size-10 shrink-0 items-center justify-center rounded-xl",
                  isIncome
                    ? "bg-emerald-500/10 text-emerald-700"
                    : "bg-muted text-muted-foreground",
                )}
              >
                <Icon className="size-5" aria-hidden="true" />
              </span>

              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">
                  {transaction.description ||
                    category?.name ||
                    "Transaction"}
                </p>
                <p className="truncate text-xs text-muted-foreground">
                  {category?.name ?? "Unknown category"} ·{" "}
                  {wallet?.name ?? "Unknown wallet"} ·{" "}
                  {format(
                    parseISO(transaction.transactionDate),
                    "MMM d, yyyy",
                  )}
                </p>
              </div>

              <span
                className={cn(
                  "shrink-0 text-sm font-medium tabular-nums",
                  isIncome ? "text-emerald-700" : "text-foreground",
                )}
              >
                {isIncome ? "+" : "−"}
                {formatTransactionAmount(
                  transaction.amount,
                  wallet?.currencyCode,
                )}
              </span>
            </li>
          );
        })}
      </ul>
    </article>
  );
}
