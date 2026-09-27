import { WalletCardsIcon } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

import { GoogleAuthButton } from "@/features/auth/components/google-auth-button";

type AuthPanelProps = {
  title: string;
  description: string;
  actionLabel: string;
  footer: ReactNode;
};

export function AuthPanel({
  title,
  description,
  actionLabel,
  footer,
}: AuthPanelProps) {
  return (
    <main className="w-full max-w-sm">
      <Link
        href="/"
        className="mb-8 inline-flex items-center gap-2 rounded-md font-heading font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground">
          <WalletCardsIcon className="size-5" aria-hidden="true" />
        </span>
        Finory
      </Link>

      <div className="rounded-2xl border bg-card p-6 shadow-sm sm:p-8">
        <div className="mb-7 space-y-2">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            {title}
          </h1>
          <p className="text-sm leading-6 text-muted-foreground">
            {description}
          </p>
        </div>

        <GoogleAuthButton>{actionLabel}</GoogleAuthButton>

        <div className="mt-6 text-center text-sm text-muted-foreground">
          {footer}
        </div>
      </div>
    </main>
  );
}
