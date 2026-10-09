export const maxQuantityPerLine = 100;

export interface CartLine {
  productId: number;
  sku: string;
  name: string;
  unitPriceCents: number;
  quantity: number;
  available: number;
}

export interface Cart {
  lines: CartLine[];
}

export interface CartProduct {
  id: number;
  sku: string;
  name: string;
  priceCents: number;
  stock: number;
}

export type CartAction =
  | { type: "add"; product: CartProduct; quantity?: number }
  | { type: "setQuantity"; productId: number; quantity: number }
  | { type: "remove"; productId: number }
  | { type: "clear" };

export const emptyCart: Cart = { lines: [] };

function clamp(quantity: number, available: number): number {
  const ceiling = Math.min(available, maxQuantityPerLine);
  return Math.max(1, Math.min(Math.floor(quantity), ceiling));
}

export function cartReducer(cart: Cart, action: CartAction): Cart {
  switch (action.type) {
    case "add": {
      const { product } = action;
      if (product.stock < 1) {
        return cart;
      }
      const existing = cart.lines.find((l) => l.productId === product.id);
      const wanted = (existing?.quantity ?? 0) + (action.quantity ?? 1);
      const line: CartLine = {
        productId: product.id,
        sku: product.sku,
        name: product.name,
        unitPriceCents: product.priceCents,
        available: product.stock,
        quantity: clamp(wanted, product.stock),
      };
      return existing
        ? { lines: cart.lines.map((l) => (l.productId === product.id ? line : l)) }
        : { lines: [...cart.lines, line] };
    }
    case "setQuantity":
      if (action.quantity < 1) {
        return { lines: cart.lines.filter((l) => l.productId !== action.productId) };
      }
      return {
        lines: cart.lines.map((l) =>
          l.productId === action.productId ? { ...l, quantity: clamp(action.quantity, l.available) } : l,
        ),
      };
    case "remove":
      return { lines: cart.lines.filter((l) => l.productId !== action.productId) };
    case "clear":
      return emptyCart;
  }
}

export function cartTotalCents(cart: Cart): number {
  return cart.lines.reduce((sum, l) => sum + l.unitPriceCents * l.quantity, 0);
}

export function cartItemCount(cart: Cart): number {
  return cart.lines.reduce((sum, l) => sum + l.quantity, 0);
}

export function isCart(value: unknown): value is Cart {
  if (typeof value !== "object" || value === null || !Array.isArray((value as Cart).lines)) {
    return false;
  }
  return (value as Cart).lines.every(
    (l) =>
      typeof l === "object" &&
      l !== null &&
      Number.isInteger(l.productId) &&
      Number.isInteger(l.quantity) &&
      Number.isInteger(l.unitPriceCents) &&
      typeof l.name === "string",
  );
}
