"use client";

import { format, parseISO } from "date-fns";
import { ChartNoAxesCombinedIcon } from "lucide-react";
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { useTrendSeries } from "@/features/dashboard/hooks/use-trend-series";
import { formatAmount } from "@/features/dashboard/lib/format-amount";
import { getApiErrorMessage } from "@/lib/api/client";

const incomeColor = "#047857";
const expenseColor = "var(--destructive)";

function formatAxisAmount(value: number) {
  return new Intl.NumberFormat(undefined, {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(value);
}

function formatDay(value: string) {
  return format(parseISO(value), "MMM d");
}

export function IncomeExpenseTrendChart() {
  const trendQuery = useTrendSeries();

  if (trendQuery.isPending) {
    return (
      <LoadingState
        title="Loading income and expense trend"
        description="Preparing this month's daily activity."
        className="min-h-96"
      />
    );
  }

  if (trendQuery.isError) {
    return (
      <ErrorState
        title="Income and expense trend unavailable"
        description={getApiErrorMessage(trendQuery.error, {
          defaultMessage: "We could not load this month's daily activity.",
        })}
        className="min-h-96"
        onRetry={() => trendQuery.refetch()}
      />
    );
  }

  const { month, points } = trendQuery.data;
  const monthLabel = format(parseISO(`${month}-01`), "MMMM yyyy");
  const hasActivity = points.some(
    (point) => Number(point.income) > 0 || Number(point.expense) > 0,
  );

  if (!hasActivity) {
    return (
      <EmptyState
        title="No activity this month"
        description={`Add a transaction to see income and expense trends for ${monthLabel}.`}
        icon={
          <ChartNoAxesCombinedIcon className="size-5" aria-hidden="true" />
        }
        className="min-h-96"
      />
    );
  }

  const chartData = points.map((point) => ({
    date: point.date,
    income: Number(point.income),
    expense: Number(point.expense),
  }));

  return (
    <article className="rounded-xl border bg-card p-5 shadow-sm">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h2 className="font-heading text-lg font-semibold tracking-tight">
            Income and expense trend
          </h2>
          <p className="text-sm text-muted-foreground">
            Daily financial activity for {monthLabel}.
          </p>
        </div>

        <ul className="flex gap-4 text-xs text-muted-foreground" aria-label="Chart legend">
          <li className="flex items-center gap-2">
            <span
              className="h-0.5 w-5 rounded-full"
              style={{ backgroundColor: incomeColor }}
              aria-hidden="true"
            />
            Income
          </li>
          <li className="flex items-center gap-2">
            <span
              className="h-0.5 w-5 rounded-full"
              style={{ backgroundColor: expenseColor }}
              aria-hidden="true"
            />
            Expense
          </li>
        </ul>
      </header>

      <div
        role="img"
        aria-label={`Daily income and expense trend for ${monthLabel}`}
        className="mt-6 h-80 min-w-0"
      >
        <ResponsiveContainer width="100%" height="100%">
          <LineChart
            data={chartData}
            margin={{ top: 8, right: 12, bottom: 0, left: 0 }}
            accessibilityLayer
          >
            <CartesianGrid
              vertical={false}
              stroke="var(--border)"
              strokeDasharray="3 3"
            />
            <XAxis
              dataKey="date"
              tickFormatter={formatDay}
              tickLine={false}
              axisLine={false}
              minTickGap={28}
              tick={{ fill: "var(--muted-foreground)", fontSize: 12 }}
            />
            <YAxis
              tickFormatter={formatAxisAmount}
              tickLine={false}
              axisLine={false}
              width={56}
              tick={{ fill: "var(--muted-foreground)", fontSize: 12 }}
            />
            <Tooltip
              labelFormatter={(label) => formatDay(String(label))}
              formatter={(value, name) => [
                formatAmount(String(value)),
                name === "income" ? "Income" : "Expense",
              ]}
              contentStyle={{
                backgroundColor: "var(--popover)",
                border: "1px solid var(--border)",
                borderRadius: "var(--radius)",
                color: "var(--popover-foreground)",
              }}
            />
            <Line
              type="monotone"
              dataKey="income"
              stroke={incomeColor}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4 }}
              isAnimationActive={false}
            />
            <Line
              type="monotone"
              dataKey="expense"
              stroke={expenseColor}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4 }}
              isAnimationActive={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </article>
  );
}
