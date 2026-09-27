import type { z } from "zod";

import type {
  categorySchema,
  categoryTypeSchema,
  createCategorySchema,
  updateCategorySchema,
} from "@/features/categories/schemas/category-schema";

export type Category = z.infer<typeof categorySchema>;
export type CategoryType = z.infer<typeof categoryTypeSchema>;
export type CreateCategoryInput = z.infer<typeof createCategorySchema>;
export type UpdateCategoryInput = z.infer<typeof updateCategorySchema>;
