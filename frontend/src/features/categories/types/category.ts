import type { z } from "zod";

import type {
  categorySchema,
  categoryTypeSchema,
} from "@/features/categories/schemas/category-schema";

export type Category = z.infer<typeof categorySchema>;
export type CategoryType = z.infer<typeof categoryTypeSchema>;
