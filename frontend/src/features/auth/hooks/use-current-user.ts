"use client";

import { useQuery } from "@tanstack/react-query";

import {
  authQueryKeys,
  getCurrentUser,
} from "@/features/auth/api/auth-api";

export function useCurrentUser() {
  return useQuery({
    queryKey: authQueryKeys.currentUser,
    queryFn: getCurrentUser,
    retry: false,
  });
}
