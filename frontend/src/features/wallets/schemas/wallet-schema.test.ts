import assert from "node:assert/strict";
import test from "node:test";

import { createWalletSchema } from "./wallet-schema.ts";

const validWallet = {
  name: "  Daily Cash  ",
  type: "cash" as const,
  balance: "-25.1250",
  currencyCode: "THB",
};

test("wallet input accepts signed precise balances and normalizes text", () => {
  assert.deepEqual(createWalletSchema.parse(validWallet), {
    ...validWallet,
    name: "Daily Cash",
  });
});

test("wallet input rejects malformed balances and currency codes", () => {
  for (const balance of ["01", "1.12345", "1000000000000000", "NaN"]) {
    assert.equal(
      createWalletSchema.safeParse({ ...validWallet, balance }).success,
      false,
      `expected ${balance} to be rejected`,
    );
  }

  for (const currencyCode of ["thb", "US", "USDT"]) {
    assert.equal(
      createWalletSchema.safeParse({ ...validWallet, currencyCode }).success,
      false,
      `expected ${currencyCode} to be rejected`,
    );
  }
});
