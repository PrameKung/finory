import { z } from "zod";

export const transactionTypeSchema = z.enum(["income", "expense"]);

export const transactionSchema = z.object({
  id: z.uuid(),
  categoryId: z.uuid(),
  walletId: z.uuid(),
  type: transactionTypeSchema,
  amount: z.string(),
  description: z.string().nullable(),
  transactionDate: z.iso.date(),
  createdAt: z.iso.datetime(),
  updatedAt: z.iso.datetime(),
});

export const transactionsSchema = z.array(transactionSchema);

const amountPattern = /^(0|[1-9][0-9]{0,14})(\.[0-9]{1,4})?$/;

export const createTransactionSchema = z.object({
  categoryId: z.string().trim().min(1, "Select a category."),
  walletId: z.string().trim().min(1, "Select a wallet."),
  type: transactionTypeSchema,
  amount: z
    .string()
    .trim()
    .min(1, "Enter an amount.")
    .regex(
      amountPattern,
      "Enter a valid amount with up to four decimal places.",
    )
    .refine((amount) => Number(amount) > 0, "Amount must be greater than zero."),
  description: z
    .string()
    .trim()
    .max(500, "Notes must be 500 characters or fewer."),
  transactionDate: z
    .string()
    .refine(
      (date) => z.iso.date().safeParse(date).success,
      "Select a valid date.",
    ),
});
