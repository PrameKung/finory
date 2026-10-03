import { CircleAlertIcon, WalletCardsIcon } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

import { GoogleAuthButton } from "@/features/auth/components/google-auth-button";
import { cn } from "@/lib/utils";

type AuthPanelProps = {
  title: string;
  description: string;
  actionLabel: string;
  footer: ReactNode;
  appearance?: "card" | "plain";
  errorMessage?: string;
};

export function AuthPanel({
  title,
  description,
  actionLabel,
  footer,
  appearance = "card",
  errorMessage,
}: AuthPanelProps) {
  return (
    <main className="w-full max-w-md">
      <Link
        href="/"
        className={cn(
          "mb-10 inline-flex items-center gap-2.5 rounded-md font-heading text-lg font-semibold tracking-tight focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring lg:mb-12",
          appearance === "plain" && "lg:hidden",
        )}
      >
        <span className="flex size-10 items-center justify-center rounded-xl bg-[#0d1f1a] text-emerald-300 shadow-sm">
          <WalletCardsIcon className="size-5" aria-hidden="true" />
        </span>
        Finory
      </Link>

      <div
        className={
          appearance === "card"
            ? "rounded-2xl border bg-card p-6 shadow-sm sm:p-8"
            : ""
        }
      >
        <div className="mb-8 space-y-3">
          <h1 className="font-heading text-3xl font-semibold tracking-[-0.025em] sm:text-4xl">
            {title}
          </h1>
          <p className="max-w-sm text-sm leading-6 text-muted-foreground sm:text-base">
            {description}
          </p>
        </div>

        {errorMessage ? (
          <div
            role="alert"
            className="mb-5 flex gap-3 rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm leading-5 text-destructive"
          >
            <CircleAlertIcon className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
            <p>{errorMessage}</p>
          </div>
        ) : null}

        <GoogleAuthButton>{actionLabel}</GoogleAuthButton>

        <div className="mt-7 text-center text-sm text-muted-foreground">
          {footer}
        </div>
      </div>
    </main>
  );
}
