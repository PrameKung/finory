import type { z } from "zod";

import type {
  createTransactionSchema,
  transactionSchema,
  transactionTypeSchema,
  updateTransactionSchema,
} from "@/features/transactions/schemas/transaction-schema";

export type Transaction = z.infer<typeof transactionSchema>;
export type TransactionType = z.infer<typeof transactionTypeSchema>;
export type CreateTransactionInput = z.infer<typeof createTransactionSchema>;
export type UpdateTransactionInput = z.infer<typeof updateTransactionSchema>;

export type TransactionFilters = {
  month?: string;
  type?: TransactionType;
};
