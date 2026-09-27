"use client";

import { format, parseISO } from "date-fns";
import { ArrowUpRightIcon, LoaderCircleIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useMonthlySummary } from "@/features/dashboard/hooks/use-monthly-summary";

function formatAmount(amount: string) {
  const value = Number(amount);

  if (!Number.isFinite(value)) {
    return amount;
  }

  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 4,
  }).format(value);
}

export function MonthlyExpenseCard() {
  const summaryQuery = useMonthlySummary();

  if (summaryQuery.isPending) {
    return (
      <article
        aria-busy="true"
        aria-label="Loading monthly expense"
        className="min-h-40 rounded-xl border bg-card p-5 shadow-sm"
      >
        <div className="flex h-full min-h-30 items-center justify-center text-muted-foreground">
          <LoaderCircleIcon
            className="size-5 animate-spin motion-reduce:animate-none"
            aria-hidden="true"
          />
          <span className="sr-only">Loading monthly expense</span>
        </div>
      </article>
    );
  }

  if (summaryQuery.isError) {
    return (
      <article
        role="alert"
        className="flex min-h-40 flex-col justify-between gap-4 rounded-xl border border-destructive/20 bg-destructive/5 p-5 shadow-sm"
      >
        <div className="space-y-1">
          <p className="text-sm font-medium">Monthly expense unavailable</p>
          <p className="text-sm text-muted-foreground">
            We could not load this month&apos;s expense.
          </p>
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="w-fit"
          onClick={() => summaryQuery.refetch()}
        >
          Try again
        </Button>
      </article>
    );
  }

  const monthLabel = format(
    parseISO(`${summaryQuery.data.month}-01`),
    "MMMM yyyy",
  );

  return (
    <article className="min-h-40 rounded-xl border bg-card p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-1">
          <p className="text-sm font-medium text-muted-foreground">
            Monthly expense
          </p>
          <p className="font-heading text-3xl font-semibold tracking-tight tabular-nums">
            {formatAmount(summaryQuery.data.expense)}
          </p>
        </div>
        <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
          <ArrowUpRightIcon className="size-5" aria-hidden="true" />
        </span>
      </div>
      <p className="mt-6 text-sm text-muted-foreground">{monthLabel}</p>
    </article>
  );
}
