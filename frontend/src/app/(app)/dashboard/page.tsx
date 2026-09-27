import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { ExpenseDistributionChart } from "@/features/dashboard/components/expense-distribution-chart";
import { IncomeExpenseTrendChart } from "@/features/dashboard/components/income-expense-trend-chart";
import { MonthlyComparison } from "@/features/dashboard/components/monthly-comparison";
import { MonthlySummaryCard } from "@/features/dashboard/components/monthly-summary-card";
import { RecentTransactions } from "@/features/dashboard/components/recent-transactions";

export const metadata: Metadata = {
  title: "Dashboard",
};

export default function DashboardPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        title="Dashboard"
        description="Your monthly financial overview."
      />

      <section
        aria-label="Monthly summary"
        className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3"
      >
        <MonthlySummaryCard metric="income" />
        <MonthlySummaryCard metric="expense" />
        <MonthlySummaryCard metric="netBalance" />
      </section>

      <section aria-label="Expense analytics">
        <ExpenseDistributionChart />
      </section>

      <section aria-label="Income and expense analytics">
        <IncomeExpenseTrendChart />
      </section>

      <section aria-label="Monthly comparison">
        <MonthlyComparison />
      </section>

      <section aria-label="Recent activity">
        <RecentTransactions />
      </section>
    </div>
  );
}
