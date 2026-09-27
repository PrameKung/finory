import type { Metadata } from "next";
import { Suspense } from "react";

import { LoadingState } from "@/components/shared/loading-state";
import { PageHeader } from "@/components/shared/page-header";
import { CreateTransactionSheet } from "@/features/transactions/components/create-transaction-sheet";
import { TransactionList } from "@/features/transactions/components/transaction-list";

export const metadata: Metadata = {
  title: "Transactions",
};

export default function TransactionsPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        title="Transactions"
        description="Review and filter your income and expenses."
        action={<CreateTransactionSheet />}
      />

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
