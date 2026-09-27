import {
  environmentManager,
  MutationCache,
  QueryCache,
  QueryClient,
} from "@tanstack/react-query";

import { ApiError } from "./api/client.ts";

const currentUserQueryKey = ["auth", "current-user"] as const;

function isUnauthorized(error: unknown) {
  return error instanceof ApiError && error.status === 401;
}

function createQueryClient() {
  function expireSession(error: unknown) {
    if (isUnauthorized(error)) {
      queryClient.setQueryData(currentUserQueryKey, null);
    }
  }

  const queryClient = new QueryClient({
    queryCache: new QueryCache({ onError: expireSession }),
    mutationCache: new MutationCache({ onError: expireSession }),
    defaultOptions: {
      queries: {
        staleTime: 60 * 1000,
        retry: (failureCount, error) =>
          !isUnauthorized(error) && failureCount < 3,
      },
    },
  });

  return queryClient;
}

let browserQueryClient: QueryClient | undefined;

export function getQueryClient() {
  if (environmentManager.isServer()) {
    return createQueryClient();
  }

  browserQueryClient ??= createQueryClient();

  return browserQueryClient;
}
