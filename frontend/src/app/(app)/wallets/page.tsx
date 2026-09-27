import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Wallets",
};

export default function WalletsPage() {
  return (
    <section className="space-y-1">
      <h1 className="font-heading text-2xl font-semibold tracking-tight">
        Wallets
      </h1>
      <p className="text-sm text-muted-foreground">
        Manage the accounts where you keep your money.
      </p>
    </section>
  );
}
