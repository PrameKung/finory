import type { z } from "zod";

import type {
  createTransactionSchema,
  transactionSchema,
  transactionTypeSchema,
} from "@/features/transactions/schemas/transaction-schema";

export type Transaction = z.infer<typeof transactionSchema>;
export type TransactionType = z.infer<typeof transactionTypeSchema>;
export type CreateTransactionInput = z.infer<typeof createTransactionSchema>;

export type TransactionFilters = {
  month?: string;
  type?: TransactionType;
};
