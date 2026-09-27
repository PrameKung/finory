import type { Metadata } from "next";

import { CategoryList } from "@/features/categories/components/category-list";

export const metadata: Metadata = {
  title: "Categories",
};

export default function CategoriesPage() {
  return (
    <div className="space-y-8">
      <header className="space-y-1">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">
          Categories
        </h1>
        <p className="text-sm text-muted-foreground">
          Organize transactions with reusable income and expense categories.
        </p>
      </header>

      <CategoryList />
    </div>
  );
}
