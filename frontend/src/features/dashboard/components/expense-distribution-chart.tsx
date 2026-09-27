"use client";

import { format, parseISO } from "date-fns";
import { ChartPieIcon } from "lucide-react";
import {
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
} from "recharts";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { useCategoryDistribution } from "@/features/dashboard/hooks/use-category-distribution";
import { formatAmount } from "@/features/dashboard/lib/format-amount";

const fallbackColors = [
  "var(--chart-5)",
  "var(--chart-4)",
  "var(--chart-3)",
  "var(--chart-2)",
  "var(--chart-1)",
];

function safeCategoryColor(color: string | null, index: number) {
  return color && /^#[\da-f]{6}$/i.test(color)
    ? color
    : fallbackColors[index % fallbackColors.length];
}

export function ExpenseDistributionChart() {
  const distributionQuery = useCategoryDistribution();

  if (distributionQuery.isPending) {
    return (
      <LoadingState
        title="Loading expense distribution"
        description="Preparing this month's category breakdown."
        className="min-h-96"
      />
    );
  }

  if (distributionQuery.isError) {
    return (
      <ErrorState
        title="Expense distribution unavailable"
        description="We could not load this month's category breakdown."
        className="min-h-96"
        onRetry={() => distributionQuery.refetch()}
      />
    );
  }

  const { month, totalExpense, categories } = distributionQuery.data;
  const monthLabel = format(parseISO(`${month}-01`), "MMMM yyyy");

  if (categories.length === 0) {
    return (
      <EmptyState
        title="No expenses this month"
        description={`Add an expense to see your distribution for ${monthLabel}.`}
        icon={<ChartPieIcon className="size-5" aria-hidden="true" />}
        className="min-h-96"
      />
    );
  }

  const chartData = categories.map((category, index) => ({
    ...category,
    value: Number(category.amount),
    chartColor: safeCategoryColor(category.color, index),
  }));

  return (
    <article className="rounded-xl border bg-card p-5 shadow-sm">
      <header className="space-y-1">
        <h2 className="font-heading text-lg font-semibold tracking-tight">
          Expense distribution
        </h2>
        <p className="text-sm text-muted-foreground">
          Spending by category for {monthLabel}.
        </p>
      </header>

      <div className="mt-6 grid items-center gap-8 md:grid-cols-[minmax(0,1fr)_minmax(16rem,0.8fr)]">
        <div
          role="img"
          aria-label={`Expense distribution for ${monthLabel}`}
          className="relative h-72 min-w-0"
        >
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={chartData}
                dataKey="value"
                nameKey="name"
                innerRadius={72}
                outerRadius={108}
                paddingAngle={2}
                stroke="var(--card)"
                strokeWidth={2}
                isAnimationActive={false}
              >
                {chartData.map((category) => (
                  <Cell
                    key={category.categoryId}
                    fill={category.chartColor}
                  />
                ))}
              </Pie>
              <Tooltip
                formatter={(value) => formatAmount(String(value))}
                contentStyle={{
                  backgroundColor: "var(--popover)",
                  border: "1px solid var(--border)",
                  borderRadius: "var(--radius)",
                  color: "var(--popover-foreground)",
                }}
              />
            </PieChart>
          </ResponsiveContainer>

          <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center text-center">
            <span className="text-xs text-muted-foreground">Total expense</span>
            <span className="mt-1 font-heading text-xl font-semibold tabular-nums">
              {formatAmount(totalExpense)}
            </span>
          </div>
        </div>

        <div>
          <div className="space-y-1">
            <h3 className="font-heading text-base font-semibold">
              Category ranking
            </h3>
            <p className="text-xs text-muted-foreground">
              Highest spending first.
            </p>
          </div>

          <ol className="mt-4 space-y-3" aria-label="Expense category ranking">
            {chartData.map((category) => (
              <li
                key={category.categoryId}
                className="flex items-center gap-3 text-sm"
              >
                <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold tabular-nums text-muted-foreground">
                  {category.rank}
                </span>
                <span
                  className="size-2.5 shrink-0 rounded-full"
                  style={{ backgroundColor: category.chartColor }}
                  aria-hidden="true"
                />
                <span className="min-w-0 flex-1 truncate font-medium">
                  {category.name}
                </span>
                <span className="text-right tabular-nums">
                  <span className="block font-medium">
                    {formatAmount(category.amount)}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {category.percentage}%
                  </span>
                </span>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </article>
  );
}
