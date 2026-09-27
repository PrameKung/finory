import type { Metadata } from "next";

import { MonthlySummaryCard } from "@/features/dashboard/components/monthly-summary-card";

export const metadata: Metadata = {
  title: "Dashboard",
};

export default function DashboardPage() {
  return (
    <div className="space-y-8">
      <header className="space-y-1">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">
          Dashboard
        </h1>
        <p className="text-sm text-muted-foreground">
          Your monthly financial overview.
        </p>
      </header>

      <section
        aria-label="Monthly summary"
        className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4"
      >
        <MonthlySummaryCard metric="income" />
        <MonthlySummaryCard metric="expense" />
        <MonthlySummaryCard metric="netBalance" />
      </section>
    </div>
  );
}
