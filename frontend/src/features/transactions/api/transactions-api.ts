import {
  transactionSchema,
  transactionsSchema,
} from "@/features/transactions/schemas/transaction-schema";
import type {
  CreateTransactionInput,
  Transaction,
  TransactionFilters,
  UpdateTransactionInput,
} from "@/features/transactions/types/transaction";
import { apiRequest } from "@/lib/api/client";

const transactionsPath = "/api/v1/transactions";

export const transactionQueryKeys = {
  all: ["transactions"] as const,
  list: (filters: TransactionFilters) =>
    ["transactions", "list", filters] as const,
};

export async function getTransactions(
  filters: TransactionFilters = {},
): Promise<Transaction[]> {
  const searchParams = new URLSearchParams();

  if (filters.month) {
    searchParams.set("month", filters.month);
  }

  if (filters.type) {
    searchParams.set("type", filters.type);
  }

  const query = searchParams.toString();
  const data = await apiRequest<unknown>(
    query ? `${transactionsPath}?${query}` : transactionsPath,
    { cache: "no-store" },
  );

  return transactionsSchema.parse(data);
}

export async function createTransaction(
  input: CreateTransactionInput,
): Promise<Transaction> {
  const data = await apiRequest<unknown>(transactionsPath, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });

  return transactionSchema.parse(data);
}

export async function updateTransaction(
  id: string,
  input: UpdateTransactionInput,
): Promise<Transaction> {
  const data = await apiRequest<unknown>(`${transactionsPath}/${id}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });

  return transactionSchema.parse(data);
}
