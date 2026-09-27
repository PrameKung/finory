import type { ReactNode } from "react";

import { AppShell } from "@/components/layout/app-shell";
import { AuthGuard } from "@/features/auth/components/auth-guard";

type AuthenticatedLayoutProps = {
  children: ReactNode;
};

export default function AuthenticatedLayout({
  children,
}: AuthenticatedLayoutProps) {
  return (
    <AuthGuard>
      <AppShell>{children}</AppShell>
    </AuthGuard>
  );
}
