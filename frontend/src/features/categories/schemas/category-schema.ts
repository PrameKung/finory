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
