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

test("apiRequest supports successful responses without a body", async () => {
  globalThis.fetch = async () => new Response(null, { status: 204 });

  assert.equal(
    await apiRequest<void>("/api/v1/transactions/transaction-id", {
      method: "DELETE",
    }),
    undefined,
  );
});

test("apiRequest normalizes network failures but preserves abort errors", async () => {
  const networkFailure = new TypeError("fetch failed");
  globalThis.fetch = async () => {
    throw networkFailure;
  };

  await assert.rejects(
    apiRequest("/api/v1/transactions"),
    (error) =>
      error instanceof ApiError &&
      error.status === 0 &&
      error.code === "network_error" &&
      error.cause === networkFailure,
  );

  const controller = new AbortController();
  const abortFailure = new DOMException("The operation was aborted", "AbortError");
  controller.abort();
  globalThis.fetch = async () => {
    throw abortFailure;
  };

  await assert.rejects(
    apiRequest("/api/v1/transactions", { signal: controller.signal }),
    (error) => error === abortFailure,
  );
});

test("apiRequest rejects a successful response with invalid JSON", async () => {
  globalThis.fetch = async () =>
    new Response("not-json", {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });

  await assert.rejects(
    apiRequest("/api/v1/transactions"),
    (error) =>
      error instanceof ApiError &&
      error.status === 200 &&
      error.code === "invalid_response",
  );
});
