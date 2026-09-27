import assert from "node:assert/strict";
import test from "node:test";

import {
  ApiError,
  getApiErrorMessage,
  parseApiErrorBody,
} from "./errors.ts";

test("parseApiErrorBody normalizes supported API error shapes", () => {
  assert.deepEqual(parseApiErrorBody({ error: " invalid_request " }), {
    code: "invalid_request",
    message: undefined,
  });
  assert.deepEqual(
    parseApiErrorBody({ code: "conflict", message: "Already exists" }),
    { code: "conflict", message: "Already exists" },
  );
  assert.deepEqual(
    parseApiErrorBody({
      error: { code: "unauthorized", message: "Authentication required" },
    }),
    { code: "unauthorized", message: "Authentication required" },
  );
  assert.deepEqual(parseApiErrorBody({ error: 123 }), {
    code: undefined,
    message: undefined,
  });
});

test("getApiErrorMessage prefers code and status overrides", () => {
  const options = {
    defaultMessage: "Fallback",
    codeMessages: { conflict: "Code message" },
    statusMessages: { 409: "Status message" },
  };

  assert.equal(
    getApiErrorMessage(new ApiError(409, "conflict"), options),
    "Code message",
  );
  assert.equal(
    getApiErrorMessage(new ApiError(409, "unknown"), options),
    "Status message",
  );
});

test("getApiErrorMessage provides consistent shared fallbacks", () => {
  const options = { defaultMessage: "Fallback" };

  assert.equal(
    getApiErrorMessage(new ApiError(0, "network_error"), options),
    "We could not reach the server. Check your connection and try again.",
  );
  assert.equal(
    getApiErrorMessage(new ApiError(401), options),
    "Your session has expired. Sign in and try again.",
  );
  assert.equal(
    getApiErrorMessage(new ApiError(503), options),
    "The service is temporarily unavailable. Please try again.",
  );
  assert.equal(getApiErrorMessage(new Error("unexpected"), options), "Fallback");
});
