import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { CategoryList } from "@/features/categories/components/category-list";
import { CreateCategorySheet } from "@/features/categories/components/create-category-sheet";

export const metadata: Metadata = {
  title: "Categories",
};

export default function CategoriesPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        title="Categories"
        description="Organize transactions with reusable income and expense categories."
        action={<CreateCategorySheet />}
      />

      <CategoryList />
    </div>
  );
}
