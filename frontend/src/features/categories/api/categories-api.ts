import {
  categoriesSchema,
  categorySchema,
} from "@/features/categories/schemas/category-schema";
import type {
  Category,
  CreateCategoryInput,
  UpdateCategoryInput,
} from "@/features/categories/types/category";
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

export async function createCategory(
  input: CreateCategoryInput,
): Promise<Category> {
  const data = await apiRequest<unknown>(categoriesPath, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      ...input,
      icon: "",
      color: "",
    }),
  });

  return categorySchema.parse(data);
}

export async function updateCategory(
  id: string,
  input: UpdateCategoryInput,
): Promise<Category> {
  const data = await apiRequest<unknown>(`${categoriesPath}/${id}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });

  return categorySchema.parse(data);
}

export function deleteCategory(id: string) {
  return apiRequest<void>(`${categoriesPath}/${id}`, {
    method: "DELETE",
  });
}
