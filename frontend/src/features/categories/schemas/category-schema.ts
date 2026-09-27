import { z } from "zod";

export const categoryTypeSchema = z.enum(["income", "expense"]);

export const categorySchema = z.object({
  id: z.uuid(),
  name: z.string(),
  type: categoryTypeSchema,
  icon: z.string().nullable(),
  color: z.string().nullable(),
  isDefault: z.boolean(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export const categoriesSchema = z.array(categorySchema);

export const createCategorySchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Enter a category name.")
    .max(100, "Category names must be 100 characters or fewer."),
  type: categoryTypeSchema,
});
