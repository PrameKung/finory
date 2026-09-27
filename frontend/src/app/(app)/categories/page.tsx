import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Categories",
};

export default function CategoriesPage() {
  return (
    <section className="space-y-1">
      <h1 className="font-heading text-2xl font-semibold tracking-tight">
        Categories
      </h1>
      <p className="text-sm text-muted-foreground">
        Organize transactions with reusable categories.
      </p>
    </section>
  );
}
