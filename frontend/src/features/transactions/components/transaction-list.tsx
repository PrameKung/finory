"use client";

import { format, parseISO } from "date-fns";
import {
  ArrowDownLeftIcon,
  ArrowLeftRightIcon,
  ArrowUpRightIcon,
  RotateCcwIcon,
} from "lucide-react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { Button } from "@/components/ui/button";
import { useCategories } from "@/features/categories/hooks/use-categories";
import type { Category } from "@/features/categories/types/category";
import { EditTransactionSheet } from "@/features/transactions/components/edit-transaction-sheet";
import { useTransactions } from "@/features/transactions/hooks/use-transactions";
import type {
  Transaction,
  TransactionFilters,
  TransactionType,
} from "@/features/transactions/types/transaction";
import { useWallets } from "@/features/wallets/hooks/use-wallets";
import type { Wallet } from "@/features/wallets/types/wallet";
import { cn } from "@/lib/utils";

const monthPattern = /^\d{4}-(0[1-9]|1[0-2])$/;

function getFilters(searchParams: URLSearchParams): TransactionFilters {
  const month = searchParams.get("month") ?? "";
  const type = searchParams.get("type") ?? "";

  return {
    month: monthPattern.test(month) ? month : undefined,
    type:
      type === "income" || type === "expense"
        ? (type as TransactionType)
        : undefined,
  };
}

function formatAmount(amount: string, currencyCode?: string) {
  const value = Number(amount);

  if (!Number.isFinite(value)) {
    return currencyCode ? `${amount} ${currencyCode}` : amount;
  }

  if (!currencyCode) {
    return new Intl.NumberFormat(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 4,
    }).format(value);
  }

  try {
    return new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: currencyCode,
      minimumFractionDigits: 2,
      maximumFractionDigits: 4,
    }).format(value);
  } catch {
    return `${amount} ${currencyCode}`;
  }
}

function TransactionAmount({
  transaction,
  wallet,
}: {
  transaction: Transaction;
  wallet?: Wallet;
}) {
  const isIncome = transaction.type === "income";

  return (
    <span
      className={cn(
        "font-medium whitespace-nowrap tabular-nums",
        isIncome ? "text-emerald-700" : "text-foreground",
      )}
    >
      {isIncome ? "+" : "−"}
      {formatAmount(transaction.amount, wallet?.currencyCode)}
    </span>
  );
}

function TypeIcon({ type }: { type: TransactionType }) {
  const isIncome = type === "income";
  const Icon = isIncome ? ArrowDownLeftIcon : ArrowUpRightIcon;

  return (
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
  );
}

function TransactionFiltersBar({ filters }: { filters: TransactionFilters }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const hasFilters = Boolean(filters.month || filters.type);

  function updateFilter(name: "month" | "type", value: string) {
    const nextSearchParams = new URLSearchParams(searchParams.toString());

    if (value) {
      nextSearchParams.set(name, value);
    } else {
      nextSearchParams.delete(name);
    }

    const query = nextSearchParams.toString();
    router.replace(query ? `${pathname}?${query}` : pathname, { scroll: false });
  }

  return (
    <div className="flex flex-col gap-3 rounded-xl border bg-card p-4 shadow-sm sm:flex-row sm:items-end">
      <label className="grid gap-1.5 text-sm font-medium sm:w-52">
        Month
        <input
          type="month"
          value={filters.month ?? ""}
          onChange={(event) => updateFilter("month", event.target.value)}
          className="h-9 rounded-lg border bg-background px-3 text-sm font-normal outline-none transition-shadow focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
        />
      </label>

      <label className="grid gap-1.5 text-sm font-medium sm:w-44">
        Type
        <select
          value={filters.type ?? ""}
          onChange={(event) => updateFilter("type", event.target.value)}
          className="h-9 rounded-lg border bg-background px-3 text-sm font-normal outline-none transition-shadow focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
        >
          <option value="">All types</option>
          <option value="income">Income</option>
          <option value="expense">Expense</option>
        </select>
      </label>

      {hasFilters ? (
        <Button
          type="button"
          variant="ghost"
          className="sm:ml-auto"
          onClick={() => router.replace(pathname, { scroll: false })}
        >
          <RotateCcwIcon aria-hidden="true" />
          Clear filters
        </Button>
      ) : null}
    </div>
  );
}

function TransactionTable({
  transactions,
  categoriesById,
  walletsById,
}: {
  transactions: Transaction[];
  categoriesById: Map<string, Category>;
  walletsById: Map<string, Wallet>;
}) {
  return (
    <div className="overflow-hidden rounded-xl border bg-card shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full min-w-180 text-left text-sm">
          <thead className="border-b bg-muted/50 text-xs font-medium tracking-wide text-muted-foreground uppercase">
            <tr>
              <th scope="col" className="px-5 py-3">Transaction</th>
              <th scope="col" className="px-5 py-3">Category</th>
              <th scope="col" className="px-5 py-3">Wallet</th>
              <th scope="col" className="px-5 py-3">Date</th>
              <th scope="col" className="px-5 py-3 text-right">Amount</th>
              <th scope="col" className="w-12 px-3 py-3">
                <span className="sr-only">Actions</span>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {transactions.map((transaction) => {
              const category = categoriesById.get(transaction.categoryId);
              const wallet = walletsById.get(transaction.walletId);

              return (
                <tr key={transaction.id} className="transition-colors hover:bg-muted/30">
                  <td className="px-5 py-4">
                    <div className="flex items-center gap-3">
                      <TypeIcon type={transaction.type} />
                      <div className="min-w-0">
                        <p className="max-w-72 truncate font-medium">
                          {transaction.description || category?.name || "Transaction"}
                        </p>
                        <p className="text-xs capitalize text-muted-foreground">
                          {transaction.type}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="px-5 py-4 text-muted-foreground">
                    {category?.name ?? "Unknown category"}
                  </td>
                  <td className="px-5 py-4 text-muted-foreground">
                    {wallet?.name ?? "Unknown wallet"}
                  </td>
                  <td className="px-5 py-4 whitespace-nowrap text-muted-foreground">
                    {format(parseISO(transaction.transactionDate), "MMM d, yyyy")}
                  </td>
                  <td className="px-5 py-4 text-right">
                    <TransactionAmount transaction={transaction} wallet={wallet} />
                  </td>
                  <td className="px-3 py-4 text-right">
                    <EditTransactionSheet transaction={transaction} />
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function TransactionCards({
  transactions,
  categoriesById,
  walletsById,
}: {
  transactions: Transaction[];
  categoriesById: Map<string, Category>;
  walletsById: Map<string, Wallet>;
}) {
  return (
    <ul className="divide-y overflow-hidden rounded-xl border bg-card shadow-sm md:hidden">
      {transactions.map((transaction) => {
        const category = categoriesById.get(transaction.categoryId);
        const wallet = walletsById.get(transaction.walletId);

        return (
          <li key={transaction.id} className="flex items-center gap-3 p-4">
            <TypeIcon type={transaction.type} />
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">
                {transaction.description || category?.name || "Transaction"}
              </p>
              <p className="truncate text-xs text-muted-foreground">
                {category?.name ?? "Unknown category"} · {wallet?.name ?? "Unknown wallet"}
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                {format(parseISO(transaction.transactionDate), "MMM d, yyyy")}
              </p>
            </div>
            <div className="flex shrink-0 flex-col items-end gap-1 text-right text-sm">
              <TransactionAmount transaction={transaction} wallet={wallet} />
              <EditTransactionSheet transaction={transaction} />
            </div>
          </li>
        );
      })}
    </ul>
  );
}

export function TransactionList() {
  const searchParams = useSearchParams();
  const filters = getFilters(searchParams);
  const transactionsQuery = useTransactions(filters);
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

  let content;

  if (isPending) {
    content = (
      <LoadingState
        title="Loading transactions"
        description="Fetching your income and expense history."
      />
    );
  } else if (isError) {
    content = (
      <ErrorState
        title="Could not load transactions"
        description="Your transaction history is temporarily unavailable. Please try again."
        onRetry={retry}
      />
    );
  } else if (transactionsQuery.data.length === 0) {
    content = (
      <EmptyState
        title={filters.month || filters.type ? "No matching transactions" : "No transactions yet"}
        description={
          filters.month || filters.type
            ? "Try changing or clearing your filters."
            : "Your income and expenses will appear here once you add them."
        }
        icon={<ArrowLeftRightIcon className="size-5" aria-hidden="true" />}
      />
    );
  } else {
    const categoriesById = new Map(
      categoriesQuery.data.map((category) => [category.id, category]),
    );
    const walletsById = new Map(
      walletsQuery.data.map((wallet) => [wallet.id, wallet]),
    );

    content = (
      <section aria-labelledby="transaction-list-heading">
        <div className="mb-3 flex items-end justify-between gap-4">
          <div>
            <h2
              id="transaction-list-heading"
              className="font-heading text-lg font-semibold tracking-tight"
            >
              Transaction history
            </h2>
            <p className="text-sm text-muted-foreground">
              Newest transactions appear first.
            </p>
          </div>
          <span className="shrink-0 text-sm tabular-nums text-muted-foreground">
            {transactionsQuery.data.length}{" "}
            {transactionsQuery.data.length === 1 ? "transaction" : "transactions"}
          </span>
        </div>

        <div className="hidden md:block">
          <TransactionTable
            transactions={transactionsQuery.data}
            categoriesById={categoriesById}
            walletsById={walletsById}
          />
        </div>
        <TransactionCards
          transactions={transactionsQuery.data}
          categoriesById={categoriesById}
          walletsById={walletsById}
        />
      </section>
    );
  }

  return (
    <div className="space-y-6">
      <TransactionFiltersBar filters={filters} />
      {content}
    </div>
  );
}
