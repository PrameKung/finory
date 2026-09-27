import { ArrowRightIcon, WalletCardsIcon } from "lucide-react";
import Link from "next/link";

import { Button } from "@/components/ui/button";

export default function HomePage() {
  return (
    <main className="flex min-h-svh items-center justify-center bg-muted/30 px-6 py-16">
      <div className="flex max-w-xl flex-col items-center text-center">
        <span className="mb-6 flex size-12 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-sm">
          <WalletCardsIcon className="size-6" aria-hidden="true" />
        </span>
        <p className="mb-3 text-sm font-medium text-muted-foreground">
          Finory
        </p>
        <h1 className="font-heading text-4xl font-semibold tracking-tight sm:text-5xl">
          Your finances, made clear.
        </h1>
        <p className="mt-4 max-w-md text-base leading-7 text-muted-foreground">
          Track income, expenses, wallets, and monthly progress in one simple
          place.
        </p>
        <Button asChild size="lg" className="mt-8">
          <Link href="/dashboard">
            Open dashboard
            <ArrowRightIcon data-icon="inline-end" aria-hidden="true" />
          </Link>
        </Button>
      </div>
    </main>
  );
}
