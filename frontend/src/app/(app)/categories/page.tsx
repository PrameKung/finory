import type { Metadata } from "next";

import { CategoryList } from "@/features/categories/components/category-list";
import { CreateCategorySheet } from "@/features/categories/components/create-category-sheet";

export const metadata: Metadata = {
  title: "Categories",
};

export default function CategoriesPage() {
  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            Categories
          </h1>
          <p className="text-sm text-muted-foreground">
            Organize transactions with reusable income and expense categories.
          </p>
        </div>
        <CreateCategorySheet />
      </header>

      <CategoryList />
    </div>
  );
}
