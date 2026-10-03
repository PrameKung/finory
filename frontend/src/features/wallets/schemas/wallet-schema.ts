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

const walletFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Enter a wallet name.")
    .max(100, "Wallet names must be 100 characters or fewer."),
  type: walletTypeSchema,
  balance: z
    .string()
    .trim()
    .min(1, "Enter a balance.")
    .regex(
      /^-?(0|[1-9][0-9]{0,14})(\.[0-9]{1,4})?$/,
      "Enter a valid balance with up to four decimal places.",
    ),
  currencyCode: z
    .string()
    .trim()
    .regex(/^[A-Z]{3}$/, "Enter a three-letter currency code."),
});

export const createWalletSchema = walletFormSchema;
export const updateWalletSchema = walletFormSchema;
