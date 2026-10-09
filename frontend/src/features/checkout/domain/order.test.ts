import { describe, expect, it } from "vitest";
import { fingerprint, validateCardNumber } from "./order";

describe("validateCardNumber", () => {
  it("accepts spaced card numbers", () => {
    expect(validateCardNumber("4242 4242 4242 4242")).toBeNull();
  });

  it.each(["", "4242", "4242-abcd-4242-4242"])("rejects %s", (input) => {
    expect(validateCardNumber(input)).not.toBeNull();
  });
});

describe("fingerprint", () => {
  it("ignores line order and card formatting", () => {
    const a = fingerprint({ lines: [{ productId: 2, quantity: 1 }, { productId: 1, quantity: 3 }], cardNumber: "4242 4242 4242 4242" });
    const b = fingerprint({ lines: [{ productId: 1, quantity: 3 }, { productId: 2, quantity: 1 }], cardNumber: "4242424242424242" });
    expect(a).toBe(b);
  });

  it("changes when the quantity changes", () => {
    const a = fingerprint({ lines: [{ productId: 1, quantity: 1 }], cardNumber: "4242424242424242" });
    const b = fingerprint({ lines: [{ productId: 1, quantity: 2 }], cardNumber: "4242424242424242" });
    expect(a).not.toBe(b);
  });
});
