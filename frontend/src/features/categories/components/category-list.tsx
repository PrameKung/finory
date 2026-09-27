"use client";

import {
  BriefcaseBusinessIcon,
  CarIcon,
  ChartNoAxesCombinedIcon,
  CircleMinusIcon,
  CirclePlusIcon,
  Gamepad2Icon,
  HeartPulseIcon,
  HouseIcon,
  LaptopIcon,
  ReceiptTextIcon,
  ShoppingBagIcon,
  TagIcon,
  UtensilsIcon,
  type LucideIcon,
} from "lucide-react";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { useCategories } from "@/features/categories/hooks/use-categories";
import type {
  Category,
  CategoryType,
} from "@/features/categories/types/category";

const categoryIcons: Record<string, LucideIcon> = {
  "briefcase-business": BriefcaseBusinessIcon,
  car: CarIcon,
  "chart-no-axes-combined": ChartNoAxesCombinedIcon,
  "circle-minus": CircleMinusIcon,
  "circle-plus": CirclePlusIcon,
  "gamepad-2": Gamepad2Icon,
  "heart-pulse": HeartPulseIcon,
  house: HouseIcon,
  laptop: LaptopIcon,
  "receipt-text": ReceiptTextIcon,
  "shopping-bag": ShoppingBagIcon,
  utensils: UtensilsIcon,
};

const sectionDetails: Record<
  CategoryType,
  { title: string; description: string }
> = {
  income: {
    title: "Income",
    description: "Categories for money coming in.",
  },
  expense: {
    title: "Expenses",
    description: "Categories for money going out.",
  },
};

function safeCategoryColor(color: string | null) {
  return color && /^#[\da-f]{6}$/i.test(color) ? color : undefined;
}

function CategoryItem({ category }: { category: Category }) {
  const Icon = category.icon ? categoryIcons[category.icon] : undefined;
  const color = safeCategoryColor(category.color);

  return (
    <li className="flex min-w-0 items-center gap-3 px-4 py-3.5 sm:px-5">
      <span
        className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground"
        style={color ? { backgroundColor: `${color}18`, color } : undefined}
      >
        {Icon ? (
          <Icon className="size-5" aria-hidden="true" />
        ) : (
          <TagIcon className="size-5" aria-hidden="true" />
        )}
      </span>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{category.name}</p>
        <p className="text-xs text-muted-foreground">
          {category.isDefault ? "Default category" : "Custom category"}
        </p>
      </div>
    </li>
  );
}

function CategorySection({
  type,
  categories,
}: {
  type: CategoryType;
  categories: Category[];
}) {
  const details = sectionDetails[type];

  return (
    <section aria-labelledby={`${type}-categories-heading`}>
      <div className="mb-3 flex items-end justify-between gap-4">
        <div>
          <h2
            id={`${type}-categories-heading`}
            className="font-heading text-lg font-semibold tracking-tight"
          >
            {details.title}
          </h2>
          <p className="text-sm text-muted-foreground">{details.description}</p>
        </div>
        <span className="shrink-0 text-sm tabular-nums text-muted-foreground">
          {categories.length} {categories.length === 1 ? "category" : "categories"}
        </span>
      </div>

      {categories.length === 0 ? (
        <EmptyState
          className="min-h-52"
          title={`No ${details.title.toLowerCase()} categories`}
          description={`Create a category to organize your ${type} transactions.`}
          icon={<TagIcon className="size-5" aria-hidden="true" />}
        />
      ) : (
        <ul className="divide-y overflow-hidden rounded-xl border bg-card shadow-sm">
          {categories.map((category) => (
            <CategoryItem key={category.id} category={category} />
          ))}
        </ul>
      )}
    </section>
  );
}

export function CategoryList() {
  const categoriesQuery = useCategories();

  if (categoriesQuery.isPending) {
    return (
      <LoadingState
        title="Loading categories"
        description="Fetching your income and expense categories."
      />
    );
  }

  if (categoriesQuery.isError) {
    return (
      <ErrorState
        title="Could not load categories"
        description="Your categories are temporarily unavailable. Please try again."
        onRetry={() => void categoriesQuery.refetch()}
      />
    );
  }

  const incomeCategories = categoriesQuery.data.filter(
    (category) => category.type === "income",
  );
  const expenseCategories = categoriesQuery.data.filter(
    (category) => category.type === "expense",
  );

  return (
    <div className="grid gap-8 xl:grid-cols-2">
      <CategorySection type="income" categories={incomeCategories} />
      <CategorySection type="expense" categories={expenseCategories} />
    </div>
  );
}
