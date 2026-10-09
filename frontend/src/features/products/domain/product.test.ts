import { describe, expect, it } from "vitest";
import { emptyForm, toDraft, toFormErrors, toFormValues, type ProductFormValues } from "./product";

const valid: ProductFormValues = {
  sku: "RS-001",
  name: " Running Shoes ",
  description: "Daily trainer",
  category: "Footwear",
  price: "89.99",
  stock: "150",
  weightKg: "0.35",
};

describe("toDraft", () => {
  it("builds a draft in cents and grams", () => {
    const { draft, errors } = toDraft(valid);
    expect(errors).toEqual({});
    expect(draft).toEqual({
      sku: "RS-001",
      name: "Running Shoes",
      description: "Daily trainer",
      category: "Footwear",
      priceCents: 8999,
      stock: 150,
      weightGrams: 350,
    });
  });

  it("allows an empty weight", () => {
    expect(toDraft({ ...valid, weightKg: "" }).draft?.weightGrams).toBeNull();
  });

  it("reports every invalid field", () => {
    const { draft, errors } = toDraft({ ...emptyForm, sku: "RS 001", price: "free", stock: "-5", weightKg: "heavy" });
    expect(draft).toBeNull();
    expect(Object.keys(errors).sort()).toEqual(["category", "name", "price", "sku", "stock", "weightKg"]);
  });
});

describe("form mapping", () => {
  it("round trips a product into form values", () => {
    const values = toFormValues({ id: 1, ...toDraft(valid).draft! });
    expect(values.price).toBe("89.99");
    expect(values.weightKg).toBe("0.350");
  });

  it("maps API field names to form field names", () => {
    expect(toFormErrors({ priceCents: "too high", name: "is required" })).toEqual({ price: "too high", name: "is required" });
  });
});
