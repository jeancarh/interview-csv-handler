import { describe, expect, it } from "vitest";
import { formatCents, parseFixed, toFixedString } from "./money";

describe("parseFixed", () => {
  it.each([
    ["89.99", 2, 8999],
    ["22", 2, 2200],
    ["0.5", 2, 50],
    [" 18.75 ", 2, 1875],
    ["0.35", 3, 350],
  ])("parses %s with scale %i", (input, scale, expected) => {
    expect(parseFixed(input, scale)).toBe(expected);
  });

  it.each(["free", "", "$29.99", "-1", "1.999", "1e3", ".5"])("rejects %s", (input) => {
    expect(parseFixed(input, 2)).toBeNull();
  });
});

describe("formatting", () => {
  it("formats cents as dollars", () => {
    expect(formatCents(8999)).toBe("$89.99");
    expect(formatCents(0)).toBe("$0.00");
  });

  it("round trips fixed values", () => {
    expect(toFixedString(8999, 2)).toBe("89.99");
    expect(toFixedString(350, 3)).toBe("0.350");
  });
});
