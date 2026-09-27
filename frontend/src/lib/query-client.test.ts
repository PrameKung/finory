import assert from "node:assert/strict";
import test from "node:test";

import { ApiError } from "./api/client.ts";
import { getQueryClient } from "./query-client.ts";

test("a terminal unauthorized query expires the cached session without retrying", async () => {
  const queryClient = getQueryClient();
  const currentUserQueryKey = ["auth", "current-user"] as const;
  let requests = 0;

  queryClient.setQueryData(currentUserQueryKey, { id: "user-id" });

  await assert.rejects(
    queryClient.fetchQuery({
      queryKey: ["protected-resource"],
      queryFn: async () => {
        requests += 1;
        throw new ApiError(401, "unauthorized");
      },
    }),
    (error) => error instanceof ApiError && error.status === 401,
  );

  assert.equal(requests, 1);
  assert.equal(queryClient.getQueryData(currentUserQueryKey), null);
});
