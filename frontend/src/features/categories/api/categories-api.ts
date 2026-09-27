import { categoriesSchema } from "@/features/categories/schemas/category-schema";
import type { Category } from "@/features/categories/types/category";
import { apiRequest } from "@/lib/api/client";

const categoriesPath = "/api/v1/categories";

export const categoryQueryKeys = {
  all: ["categories"] as const,
  list: ["categories", "list"] as const,
};

export async function getCategories(): Promise<Category[]> {
  const data = await apiRequest<unknown>(categoriesPath, {
    cache: "no-store",
  });

  return categoriesSchema.parse(data);
}
