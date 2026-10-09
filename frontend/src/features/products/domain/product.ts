import { parseFixed, toFixedString } from "../../../shared/money/money";
import type { FieldErrors } from "../../../shared/validation";

export interface Product {
  id: number;
  sku: string;
  name: string;
  description: string;
  category: string;
  priceCents: number;
  stock: number;
  weightGrams: number | null;
}

export type ProductDraft = Omit<Product, "id">;

export interface ProductFormValues {
  sku: string;
  name: string;
  description: string;
  category: string;
  price: string;
  stock: string;
  weightKg: string;
}

export const emptyForm: ProductFormValues = {
  sku: "",
  name: "",
  description: "",
  category: "",
  price: "",
  stock: "0",
  weightKg: "",
};

const skuPattern = /^[A-Za-z0-9][A-Za-z0-9_-]*$/;

export function toFormValues(p: Product): ProductFormValues {
  return {
    sku: p.sku,
    name: p.name,
    description: p.description,
    category: p.category,
    price: toFixedString(p.priceCents, 2),
    stock: String(p.stock),
    weightKg: p.weightGrams === null ? "" : toFixedString(p.weightGrams, 3),
  };
}

export function toDraft(values: ProductFormValues): { draft: ProductDraft | null; errors: FieldErrors } {
  const errors: FieldErrors = {};
  const sku = values.sku.trim();
  const name = values.name.trim();
  const category = values.category.trim();

  if (!name) errors.name = "Name is required";
  if (!category) errors.category = "Category is required";
  if (!sku) errors.sku = "SKU is required";
  else if (sku.length > 64 || !skuPattern.test(sku)) errors.sku = "Use letters, digits, dashes or underscores";

  const priceCents = parseFixed(values.price, 2);
  if (priceCents === null) errors.price = "Enter an amount like 19.99";

  const stock = /^\d+$/.test(values.stock.trim()) ? Number(values.stock) : null;
  if (stock === null || stock > 1_000_000) errors.stock = "Enter a whole number between 0 and 1,000,000";

  let weightGrams: number | null = null;
  if (values.weightKg.trim() !== "") {
    weightGrams = parseFixed(values.weightKg, 3);
    if (weightGrams === null) errors.weightKg = "Enter a weight like 0.35";
  }

  if (Object.keys(errors).length > 0 || priceCents === null || stock === null) {
    return { draft: null, errors };
  }
  return {
    draft: { sku, name, description: values.description.trim(), category, priceCents, stock, weightGrams },
    errors,
  };
}

const formFieldFor: Record<string, keyof ProductFormValues> = {
  priceCents: "price",
  weightGrams: "weightKg",
};

export function toFormErrors(apiFields: FieldErrors): FieldErrors {
  const out: FieldErrors = {};
  for (const [field, message] of Object.entries(apiFields)) {
    out[formFieldFor[field] ?? field] = message;
  }
  return out;
}
