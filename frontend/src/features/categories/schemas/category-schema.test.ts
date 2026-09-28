import assert from "node:assert/strict";
import test from "node:test";

import { createCategorySchema } from "./category-schema.ts";

test("category input is trimmed before submission", () => {
  assert.deepEqual(
    createCategorySchema.parse({ name: "  Groceries  ", type: "expense" }),
    { name: "Groceries", type: "expense" },
  );
});

test("category input requires a supported type and a useful name", () => {
  assert.equal(
    createCategorySchema.safeParse({ name: "   ", type: "expense" }).success,
    false,
  );
  assert.equal(
    createCategorySchema.safeParse({ name: "Salary", type: "transfer" })
      .success,
    false,
  );
  assert.equal(
    createCategorySchema.safeParse({
      name: "a".repeat(101),
      type: "income",
    }).success,
    false,
  );
});
