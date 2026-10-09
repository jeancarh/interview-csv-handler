import { describe, expect, it } from "vitest";
import { cartItemCount, cartReducer, cartTotalCents, emptyCart, isCart, type CartProduct } from "./cart";

const shoes: CartProduct = { id: 1, sku: "RS-001", name: "Running Shoes", priceCents: 8999, stock: 3 };
const mug: CartProduct = { id: 2, sku: "TM-029", name: "Travel Mug", priceCents: 1699, stock: 500 };

describe("cartReducer", () => {
  it("adds and merges lines", () => {
    let cart = cartReducer(emptyCart, { type: "add", product: shoes });
    cart = cartReducer(cart, { type: "add", product: shoes });
    cart = cartReducer(cart, { type: "add", product: mug });
    expect(cart.lines).toHaveLength(2);
    expect(cart.lines[0]?.quantity).toBe(2);
  });

  it("never goes over the available stock", () => {
    const cart = cartReducer(emptyCart, { type: "add", product: shoes, quantity: 10 });
    expect(cart.lines[0]?.quantity).toBe(3);
  });

  it("caps a line at 100 units", () => {
    const cart = cartReducer(emptyCart, { type: "add", product: mug, quantity: 400 });
    expect(cart.lines[0]?.quantity).toBe(100);
  });

  it("ignores products without stock", () => {
    expect(cartReducer(emptyCart, { type: "add", product: { ...shoes, stock: 0 } })).toBe(emptyCart);
  });

  it("removes a line when quantity drops to zero", () => {
    const cart = cartReducer(cartReducer(emptyCart, { type: "add", product: shoes }), {
      type: "setQuantity",
      productId: 1,
      quantity: 0,
    });
    expect(cart.lines).toHaveLength(0);
  });

  it("computes totals", () => {
    let cart = cartReducer(emptyCart, { type: "add", product: shoes, quantity: 2 });
    cart = cartReducer(cart, { type: "add", product: mug });
    expect(cartTotalCents(cart)).toBe(2 * 8999 + 1699);
    expect(cartItemCount(cart)).toBe(3);
  });
});

describe("isCart", () => {
  it("rejects junk from storage", () => {
    expect(isCart(null)).toBe(false);
    expect(isCart({ lines: [{ productId: "1" }] })).toBe(false);
    expect(isCart(cartReducer(emptyCart, { type: "add", product: shoes }))).toBe(true);
  });
});
