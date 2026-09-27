import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Transactions",
};

export default function TransactionsPage() {
  return (
    <section className="space-y-1">
      <h1 className="font-heading text-2xl font-semibold tracking-tight">
        Transactions
      </h1>
      <p className="text-sm text-muted-foreground">
        Review and manage your income and expenses.
      </p>
    </section>
  );
}
