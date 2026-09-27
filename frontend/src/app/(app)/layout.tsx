import type { ReactNode } from "react";

type AuthenticatedLayoutProps = {
  children: ReactNode;
};

export default function AuthenticatedLayout({
  children,
}: AuthenticatedLayoutProps) {
  return (
    <div className="flex min-h-svh w-full bg-muted/30">
      <main className="flex min-w-0 flex-1 flex-col">{children}</main>
    </div>
  );
}
