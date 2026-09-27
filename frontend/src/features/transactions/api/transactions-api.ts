import { transactionsSchema } from "@/features/transactions/schemas/transaction-schema";
import type {
  Transaction,
  TransactionFilters,
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
