import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Dashboard",
};

export default function DashboardPage() {
  return (
    <section className="space-y-1">
      <h1 className="font-heading text-2xl font-semibold tracking-tight">
        Dashboard
      </h1>
      <p className="text-sm text-muted-foreground">
        Your monthly financial overview.
      </p>
    </section>
  );
}
