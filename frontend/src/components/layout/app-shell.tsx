"use client";

import {
  ArrowLeftRightIcon,
  LayoutDashboardIcon,
  MenuIcon,
  SettingsIcon,
  TagsIcon,
  WalletCardsIcon,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";

const navigationItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboardIcon },
  { href: "/transactions", label: "Transactions", icon: ArrowLeftRightIcon },
  { href: "/categories", label: "Categories", icon: TagsIcon },
  { href: "/wallets", label: "Wallets", icon: WalletCardsIcon },
  { href: "/settings", label: "Settings", icon: SettingsIcon },
] as const;

type AppShellProps = {
  children: React.ReactNode;
};

type NavigationProps = {
  pathname: string;
  onNavigate?: () => void;
};

function Brand({ onNavigate }: { onNavigate?: () => void }) {
  return (
    <Link
      href="/dashboard"
      onClick={onNavigate}
      className="inline-flex items-center gap-2.5 rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
    >
      <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
        <WalletCardsIcon className="size-5" aria-hidden="true" />
      </span>
      <span className="font-heading text-lg font-semibold tracking-tight">
        Finory
      </span>
    </Link>
  );
}

function Navigation({ pathname, onNavigate }: NavigationProps) {
  return (
    <nav aria-label="Primary navigation" className="flex flex-col gap-1">
      {navigationItems.map(({ href, label, icon: Icon }) => {
        const isActive = pathname === href || pathname.startsWith(`${href}/`);

        return (
          <Link
            key={href}
            href={href}
            aria-current={isActive ? "page" : undefined}
            onClick={onNavigate}
            className={cn(
              "flex h-10 items-center gap-3 rounded-lg px-3 text-sm font-medium text-sidebar-foreground/70 transition-colors outline-none hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-2 focus-visible:ring-sidebar-ring",
              isActive &&
                "bg-sidebar-accent text-sidebar-accent-foreground shadow-sm",
            )}
          >
            <Icon className="size-4.5" aria-hidden="true" />
            {label}
          </Link>
        );
      })}
    </nav>
  );
}

export function AppShell({ children }: AppShellProps) {
  const pathname = usePathname();
  const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false);

  return (
    <div className="flex min-h-svh w-full bg-muted/30">
      <a
        href="#main-content"
        className="fixed top-3 left-3 z-100 -translate-y-20 rounded-md bg-background px-3 py-2 text-sm font-medium shadow-lg transition-transform focus:translate-y-0"
      >
        Skip to content
      </a>

      <aside className="sticky top-0 hidden h-svh w-64 shrink-0 flex-col border-r bg-sidebar text-sidebar-foreground lg:flex">
        <div className="flex h-18 items-center border-b px-5">
          <Brand />
        </div>
        <div className="flex-1 overflow-y-auto p-3">
          <Navigation pathname={pathname} />
        </div>
        <div className="border-t px-5 py-4 text-xs text-sidebar-foreground/60">
          Personal finance, made clear.
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-40 flex h-16 items-center gap-3 border-b bg-background/95 px-4 backdrop-blur-sm supports-backdrop-filter:bg-background/80 lg:hidden">
          <Sheet
            open={mobileNavigationOpen}
            onOpenChange={setMobileNavigationOpen}
          >
            <SheetTrigger asChild>
              <Button variant="outline" size="icon" aria-label="Open navigation">
                <MenuIcon aria-hidden="true" />
              </Button>
            </SheetTrigger>
            <SheetContent
              side="left"
              className="w-[min(20rem,85vw)] gap-0 p-0"
            >
              <SheetHeader className="border-b px-5 py-4 text-left">
                <Brand onNavigate={() => setMobileNavigationOpen(false)} />
                <SheetTitle className="sr-only">Navigation</SheetTitle>
                <SheetDescription className="sr-only">
                  Navigate between Finory application pages.
                </SheetDescription>
              </SheetHeader>
              <div className="overflow-y-auto p-3">
                <Navigation
                  pathname={pathname}
                  onNavigate={() => setMobileNavigationOpen(false)}
                />
              </div>
            </SheetContent>
          </Sheet>
          <Brand />
        </header>

        <main
          id="main-content"
          className="mx-auto flex w-full max-w-400 flex-1 flex-col p-4 sm:p-6 lg:p-8"
        >
          {children}
        </main>
      </div>
    </div>
  );
}
