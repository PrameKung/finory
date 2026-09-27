import type { Metadata } from "next";
import { Suspense } from "react";

import { LoadingState } from "@/components/shared/loading-state";
import { CreateTransactionSheet } from "@/features/transactions/components/create-transaction-sheet";
import { TransactionList } from "@/features/transactions/components/transaction-list";

export const metadata: Metadata = {
  title: "Transactions",
};

export default function TransactionsPage() {
  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            Transactions
          </h1>
          <p className="text-sm text-muted-foreground">
            Review and filter your income and expenses.
          </p>
        </div>
        <CreateTransactionSheet />
      </header>

      <Suspense
        fallback={
          <LoadingState
            title="Loading transactions"
            description="Preparing your transaction history."
          />
        }
      >
        <TransactionList />
      </Suspense>
    </div>
  );
}
