import assert from "node:assert/strict";
import test from "node:test";

import { ApiError, apiRequest } from "./client.ts";

const originalFetch = globalThis.fetch;

test.afterEach(() => {
  globalThis.fetch = originalFetch;
});

test("apiRequest refreshes an expired session and retries once", async () => {
  let protectedRequests = 0;
  let refreshRequests = 0;

  globalThis.fetch = async (input, init) => {
    const url = String(input);
    assert.equal(init?.credentials, "include");

    if (url.endsWith("/api/v1/auth/refresh")) {
      refreshRequests += 1;
      assert.equal(init?.method, "POST");
      return new Response(null, { status: 204 });
    }

    protectedRequests += 1;
    if (protectedRequests === 1) {
      return Response.json({ error: "unauthorized" }, { status: 401 });
    }

    return Response.json({ id: "transaction-id" });
  };

  const result = await apiRequest<{ id: string }>("/api/v1/transactions");

  assert.deepEqual(result, { id: "transaction-id" });
  assert.equal(protectedRequests, 2);
  assert.equal(refreshRequests, 1);
});

test("concurrent expired requests share one refresh", async () => {
  let protectedRequests = 0;
  let refreshRequests = 0;

  globalThis.fetch = async (input) => {
    const url = String(input);

    if (url.endsWith("/api/v1/auth/refresh")) {
      refreshRequests += 1;
      await new Promise((resolve) => setImmediate(resolve));
      return new Response(null, { status: 204 });
    }

    protectedRequests += 1;
    if (protectedRequests <= 2) {
      return Response.json({ error: "unauthorized" }, { status: 401 });
    }

    return Response.json({ ok: true });
  };

  const results = await Promise.all([
    apiRequest<{ ok: boolean }>("/api/v1/categories"),
    apiRequest<{ ok: boolean }>("/api/v1/wallets"),
  ]);

  assert.deepEqual(results, [{ ok: true }, { ok: true }]);
  assert.equal(protectedRequests, 4);
  assert.equal(refreshRequests, 1);
});

test("apiRequest surfaces a terminal 401 when refresh is invalid", async () => {
  let protectedRequests = 0;
  let refreshRequests = 0;

  globalThis.fetch = async (input) => {
    const url = String(input);

    if (url.endsWith("/api/v1/auth/refresh")) {
      refreshRequests += 1;
      return Response.json({ error: "invalid_session" }, { status: 401 });
    }

    protectedRequests += 1;
    return Response.json({ error: "unauthorized" }, { status: 401 });
  };

  await assert.rejects(
    apiRequest("/api/v1/transactions"),
    (error) =>
      error instanceof ApiError &&
      error.status === 401 &&
      error.code === "unauthorized",
  );
  assert.equal(protectedRequests, 1);
  assert.equal(refreshRequests, 1);
});

test("apiRequest does not refresh repeatedly when the retried request is unauthorized", async () => {
  let protectedRequests = 0;
  let refreshRequests = 0;

  globalThis.fetch = async (input) => {
    const url = String(input);

    if (url.endsWith("/api/v1/auth/refresh")) {
      refreshRequests += 1;
      return new Response(null, { status: 204 });
    }

    protectedRequests += 1;
    return Response.json({ error: "unauthorized" }, { status: 401 });
  };

  await assert.rejects(
    apiRequest("/api/v1/transactions"),
    (error) => error instanceof ApiError && error.status === 401,
  );
  assert.equal(protectedRequests, 2);
  assert.equal(refreshRequests, 1);
});
