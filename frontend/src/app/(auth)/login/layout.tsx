import {
  ArrowDownLeftIcon,
  ArrowUpRightIcon,
  TrendingUpIcon,
  WalletCardsIcon,
} from "lucide-react";
import type { ReactNode } from "react";

type LoginLayoutProps = {
  children: ReactNode;
};

export default function LoginLayout({ children }: LoginLayoutProps) {
  return (
    <div className="grid min-h-svh bg-background lg:grid-cols-[minmax(0,1.08fr)_minmax(28rem,0.92fr)]">
      <aside className="relative hidden min-h-svh overflow-hidden bg-[#0d1f1a] px-10 py-9 text-white lg:flex lg:flex-col xl:px-16 xl:py-12">
        <div
          className="absolute -left-24 top-1/4 size-80 rounded-full bg-emerald-300/10 blur-3xl"
          aria-hidden="true"
        />
        <div
          className="absolute -right-32 -top-24 size-96 rounded-full bg-teal-200/10 blur-3xl"
          aria-hidden="true"
        />

        <div className="relative z-10 inline-flex items-center gap-3 font-heading text-lg font-semibold tracking-tight">
          <span className="flex size-10 items-center justify-center rounded-xl bg-white/10 ring-1 ring-white/15">
            <WalletCardsIcon className="size-5 text-emerald-300" aria-hidden="true" />
          </span>
          Finory
        </div>

        <div className="relative z-10 my-auto max-w-xl py-16">
          <div className="mb-8 inline-flex items-center gap-2 rounded-full border border-emerald-200/15 bg-emerald-200/10 px-3 py-1.5 text-xs font-medium text-emerald-100">
            <TrendingUpIcon className="size-3.5" aria-hidden="true" />
            Make every month count
          </div>

          <p className="max-w-lg font-heading text-4xl font-semibold leading-[1.12] tracking-[-0.035em] xl:text-5xl">
            Your money, finally in focus.
          </p>
          <p className="mt-5 max-w-md text-base leading-7 text-white/65">
            See where your money goes, build better habits, and make confident
            decisions from one calm workspace.
          </p>

          <div
            className="mt-12 max-w-lg rounded-[1.75rem] border border-white/10 bg-white/[0.07] p-5 shadow-2xl shadow-black/20 backdrop-blur-sm xl:p-6"
            aria-hidden="true"
          >
            <div className="flex items-start justify-between gap-6">
              <div>
                <p className="text-xs font-medium uppercase tracking-[0.16em] text-white/45">
                  This month
                </p>
                <p className="mt-2 text-3xl font-semibold tracking-tight">
                  ฿24,860
                </p>
                <p className="mt-1 text-xs text-emerald-300">
                  +12.4% from last month
                </p>
              </div>
              <div className="flex size-11 items-center justify-center rounded-2xl bg-emerald-300 text-[#0d1f1a]">
                <TrendingUpIcon className="size-5" />
              </div>
            </div>

            <div className="mt-7 flex h-20 items-end gap-2">
              {[32, 48, 38, 62, 52, 76, 92, 70, 86, 100, 82, 96].map(
                (height, index) => (
                  <span
                    key={`${height}-${index}`}
                    className="min-w-0 flex-1 rounded-full bg-emerald-200/25 last:bg-emerald-300"
                    style={{ height: `${height}%` }}
                  />
                ),
              )}
            </div>

            <div className="mt-6 grid grid-cols-2 gap-3 border-t border-white/10 pt-5">
              <div className="flex items-center gap-3">
                <span className="flex size-9 items-center justify-center rounded-xl bg-emerald-300/15 text-emerald-300">
                  <ArrowDownLeftIcon className="size-4" />
                </span>
                <div>
                  <p className="text-[11px] text-white/45">Income</p>
                  <p className="text-sm font-medium">฿38,200</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <span className="flex size-9 items-center justify-center rounded-xl bg-orange-300/15 text-orange-200">
                  <ArrowUpRightIcon className="size-4" />
                </span>
                <div>
                  <p className="text-[11px] text-white/45">Expenses</p>
                  <p className="text-sm font-medium">฿13,340</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <p className="relative z-10 text-xs text-white/40">
          Personal finance, made clear.
        </p>
      </aside>

      <div className="relative flex min-h-svh items-center justify-center px-6 py-12 sm:px-10 lg:px-12 xl:px-20">
        <div
          className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-border to-transparent lg:hidden"
          aria-hidden="true"
        />
        {children}
      </div>
    </div>
  );
}
