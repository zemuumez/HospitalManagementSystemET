import { test } from "node:test";
import assert from "node:assert/strict";
import { assertPhoneProof } from "./firebase-proof.ts";
test("phone bridge rejects non-phone, stale, future and malformed proofs", () => {
  const good = {
    phone_number: "+254700000001",
    auth_time: 1000,
    firebase: { sign_in_provider: "phone" },
  };
  assert.equal(assertPhoneProof(good, 1100), good.phone_number);
  assert.throws(() =>
    assertPhoneProof(
      { ...good, firebase: { sign_in_provider: "password" } },
      1100,
    ),
  );
  assert.throws(() => assertPhoneProof(good, 1400));
  assert.throws(() => assertPhoneProof(good, 900));
  assert.throws(() =>
    assertPhoneProof({ ...good, phone_number: "0700000001" }, 1100),
  );
});
