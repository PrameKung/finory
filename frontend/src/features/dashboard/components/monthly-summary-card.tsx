"use client";

import { format, parseISO } from "date-fns";
import {
  ArrowDownLeftIcon,
  ArrowUpRightIcon,
  LoaderCircleIcon,
  ScaleIcon,
  type LucideIcon,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { useMonthlySummary } from "@/features/dashboard/hooks/use-monthly-summary";
import { formatAmount } from "@/features/dashboard/lib/format-amount";
import { getApiErrorMessage } from "@/lib/api/client";
import { cn } from "@/lib/utils";

type SummaryMetric = "income" | "expense" | "netBalance";

type MetricConfig = {
  label: string;
  Icon: LucideIcon;
  iconClassName: string;
};

const metricConfig: Record<SummaryMetric, MetricConfig> = {
  income: {
    label: "Monthly income",
    Icon: ArrowDownLeftIcon,
    iconClassName: "bg-emerald-500/10 text-emerald-700",
  },
  expense: {
    label: "Monthly expense",
    Icon: ArrowUpRightIcon,
    iconClassName: "bg-destructive/10 text-destructive",
  },
  netBalance: {
    label: "Net balance",
    Icon: ScaleIcon,
    iconClassName: "bg-muted text-muted-foreground",
  },
};

export function MonthlySummaryCard({ metric }: { metric: SummaryMetric }) {
  const summaryQuery = useMonthlySummary();
  const { label, Icon, iconClassName } = metricConfig[metric];

  if (summaryQuery.isPending) {
    return (
      <article
        aria-busy="true"
        aria-label={`Loading ${label.toLowerCase()}`}
        className="min-h-40 rounded-xl border bg-card p-5 shadow-sm"
      >
        <div className="flex h-full min-h-30 items-center justify-center text-muted-foreground">
          <LoaderCircleIcon
            className="size-5 animate-spin motion-reduce:animate-none"
            aria-hidden="true"
          />
          <span className="sr-only">Loading {label.toLowerCase()}</span>
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
          <p className="text-sm font-medium">{label} unavailable</p>
          <p className="text-sm text-muted-foreground">
            {getApiErrorMessage(summaryQuery.error, {
              defaultMessage: "We could not load this month's financial summary.",
            })}
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
          <p className="text-sm font-medium text-muted-foreground">{label}</p>
          <p className="font-heading text-3xl font-semibold tracking-tight tabular-nums">
            {formatAmount(summaryQuery.data[metric])}
          </p>
        </div>
        <span
          className={cn(
            "flex size-10 shrink-0 items-center justify-center rounded-xl",
            iconClassName,
          )}
        >
          <Icon className="size-5" aria-hidden="true" />
        </span>
      </div>
      <p className="mt-6 text-sm text-muted-foreground">{monthLabel}</p>
    </article>
  );
}
