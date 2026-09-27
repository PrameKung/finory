import { z } from "zod";

export const walletTypeSchema = z.enum(["cash", "bank", "e_wallet", "other"]);

export const walletSchema = z.object({
  id: z.uuid(),
  name: z.string(),
  type: walletTypeSchema,
  balance: z.string(),
  currencyCode: z.string().regex(/^[A-Z]{3}$/),
  isDefault: z.boolean(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export const walletsSchema = z.array(walletSchema);
