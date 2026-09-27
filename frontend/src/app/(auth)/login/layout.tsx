import { WalletCardsIcon } from "lucide-react";
import type { ReactNode } from "react";

type LoginLayoutProps = {
  children: ReactNode;
};

export default function LoginLayout({ children }: LoginLayoutProps) {
  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      <aside className="relative hidden overflow-hidden bg-primary p-10 text-primary-foreground lg:flex lg:flex-col lg:justify-between">
        <div className="inline-flex items-center gap-2.5 font-heading text-lg font-semibold">
          <span className="flex size-10 items-center justify-center rounded-xl bg-primary-foreground/10 ring-1 ring-primary-foreground/15">
            <WalletCardsIcon className="size-5" aria-hidden="true" />
          </span>
          Finory
        </div>
        <div className="max-w-md space-y-4">
          <p className="font-heading text-4xl font-semibold leading-tight tracking-tight">
            A clearer view of your everyday finances.
          </p>
          <p className="text-sm leading-6 text-primary-foreground/70">
            Keep income, expenses, wallets, and monthly progress together in
            one focused workspace.
          </p>
        </div>
        <p className="text-xs text-primary-foreground/50">
          Personal finance, made clear.
        </p>
      </aside>

      <div className="flex min-h-svh items-center justify-center px-6 py-12 sm:px-10">
        {children}
      </div>
    </div>
  );
}
