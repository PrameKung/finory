import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Settings",
};

export default function SettingsPage() {
  return (
    <section className="space-y-1">
      <h1 className="font-heading text-2xl font-semibold tracking-tight">
        Settings
      </h1>
      <p className="text-sm text-muted-foreground">
        Manage your account and application preferences.
      </p>
    </section>
  );
}
