"use client";

import { format, parseISO } from "date-fns";
import {
  ArrowDownRightIcon,
  ArrowUpRightIcon,
  MinusIcon,
  type LucideIcon,
} from "lucide-react";

import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { useMonthlyComparison } from "@/features/dashboard/hooks/use-monthly-comparison";
import { formatAmount } from "@/features/dashboard/lib/format-amount";
import type { MonthlyComparison as MonthlyComparisonData } from "@/features/dashboard/types/monthly-comparison";
import { cn } from "@/lib/utils";

type ComparisonMetric = "income" | "expense";

function formatSignedAmount(amount: string) {
  return `${Number(amount) > 0 ? "+" : ""}${formatAmount(amount)}`;
}

function formatSignedPercentage(percentage: string) {
  const value = Number(percentage);
  const formatted = new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(value);

  return `${value > 0 ? "+" : ""}${formatted}%`;
}

function changePresentation(metric: ComparisonMetric, amount: string) {
  const value = Number(amount);
  let Icon: LucideIcon = MinusIcon;

  if (value > 0) {
    Icon = ArrowUpRightIcon;
  } else if (value < 0) {
    Icon = ArrowDownRightIcon;
  }

  const favorable = metric === "income" ? value > 0 : value < 0;
  const unfavorable = metric === "income" ? value < 0 : value > 0;

  return {
    Icon,
    className: cn(
      "bg-muted text-muted-foreground",
      favorable && "bg-emerald-500/10 text-emerald-700",
      unfavorable && "bg-destructive/10 text-destructive",
    ),
  };
}

function ComparisonMetricCard({
  metric,
  comparison,
  previousMonthLabel,
}: {
  metric: ComparisonMetric;
  comparison: MonthlyComparisonData;
  previousMonthLabel: string;
}) {
  const label = metric === "income" ? "Income" : "Expense";
  const change = comparison.changes[metric];
  const { Icon, className } = changePresentation(metric, change.amount);

  return (
    <section className="rounded-xl border bg-muted/20 p-4">
      <p className="text-sm font-medium text-muted-foreground">{label}</p>
      <p className="mt-1 font-heading text-2xl font-semibold tracking-tight tabular-nums">
        {formatAmount(comparison.current[metric])}
      </p>

      <div className="mt-4 flex items-end justify-between gap-4 border-t pt-4">
        <div className="text-sm">
          <p className="text-xs text-muted-foreground">{previousMonthLabel}</p>
          <p className="mt-1 font-medium tabular-nums">
            {formatAmount(comparison.previous[metric])}
          </p>
        </div>

        <div className="text-right">
          <span
            className={cn(
              "inline-flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium tabular-nums",
              className,
            )}
          >
            <Icon className="size-3.5" aria-hidden="true" />
            {change.percentage === null
              ? formatSignedAmount(change.amount)
              : formatSignedPercentage(change.percentage)}
          </span>
          <p className="mt-1 text-xs tabular-nums text-muted-foreground">
            {change.percentage === null
              ? "No previous baseline"
              : `${formatSignedAmount(change.amount)} change`}
          </p>
        </div>
      </div>
    </section>
  );
}

export function MonthlyComparison() {
  const comparisonQuery = useMonthlyComparison();

  if (comparisonQuery.isPending) {
    return (
      <LoadingState
        title="Loading monthly comparison"
        description="Comparing this month with the previous month."
        className="min-h-72"
      />
    );
  }

  if (comparisonQuery.isError) {
    return (
      <ErrorState
        title="Monthly comparison unavailable"
        description="We could not compare this month with the previous month."
        className="min-h-72"
        onRetry={() => comparisonQuery.refetch()}
      />
    );
  }

  const comparison = comparisonQuery.data;
  const currentMonthLabel = format(
    parseISO(`${comparison.month}-01`),
    "MMMM yyyy",
  );
  const previousMonthLabel = format(
    parseISO(`${comparison.previousMonth}-01`),
    "MMMM yyyy",
  );

  return (
    <article className="rounded-xl border bg-card p-5 shadow-sm">
      <header className="space-y-1">
        <h2 className="font-heading text-lg font-semibold tracking-tight">
          Month-over-month comparison
        </h2>
        <p className="text-sm text-muted-foreground">
          {currentMonthLabel} compared with {previousMonthLabel}.
        </p>
      </header>

      <div className="mt-6 grid gap-4 md:grid-cols-2">
        <ComparisonMetricCard
          metric="income"
          comparison={comparison}
          previousMonthLabel={previousMonthLabel}
        />
        <ComparisonMetricCard
          metric="expense"
          comparison={comparison}
          previousMonthLabel={previousMonthLabel}
        />
      </div>
    </article>
  );
}
