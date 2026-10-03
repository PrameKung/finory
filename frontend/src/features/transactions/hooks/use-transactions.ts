"use client";

import { useQuery } from "@tanstack/react-query";

import {
  getTransactions,
  transactionQueryKeys,
} from "@/features/transactions/api/transactions-api";
import type { TransactionFilters } from "@/features/transactions/types/transaction";

export function useTransactions(filters: TransactionFilters) {
  return useQuery({
    queryKey: transactionQueryKeys.list(filters),
    queryFn: () => getTransactions(filters),
  });
}
