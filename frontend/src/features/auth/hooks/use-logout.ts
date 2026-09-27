"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import { authQueryKeys, logout } from "@/features/auth/api/auth-api";

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: logout,
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: authQueryKeys.all });
    },
  });
}
