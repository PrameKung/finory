import assert from "node:assert/strict";
import test from "node:test";

import { createTransactionSchema } from "./transaction-schema.ts";

const validTransaction = {
  categoryId: "category-id",
  walletId: "wallet-id",
  type: "expense" as const,
  amount: "123.4567",
  description: "  Weekly groceries  ",
  transactionDate: "2026-09-28",
};

test("transaction input preserves monetary precision and normalizes text", () => {
  assert.deepEqual(createTransactionSchema.parse(validTransaction), {
    ...validTransaction,
    description: "Weekly groceries",
  });
});

test("transaction input rejects zero and malformed monetary values", () => {
  for (const amount of ["0", "00.10", "1.12345", "1000000000000000", "1e3"]) {
    assert.equal(
      createTransactionSchema.safeParse({ ...validTransaction, amount })
        .success,
      false,
      `expected ${amount} to be rejected`,
    );
  }
});

test("transaction input requires ownership references and a real calendar date", () => {
  assert.equal(
    createTransactionSchema.safeParse({
      ...validTransaction,
      categoryId: " ",
    }).success,
    false,
  );
  assert.equal(
    createTransactionSchema.safeParse({
      ...validTransaction,
      walletId: " ",
    }).success,
    false,
  );
  assert.equal(
    createTransactionSchema.safeParse({
      ...validTransaction,
      transactionDate: "2026-02-30",
    }).success,
    false,
  );
});
