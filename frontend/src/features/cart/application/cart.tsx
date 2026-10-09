import { createContext, useContext, useEffect, useMemo, useReducer, type ReactNode } from "react";
import { cartItemCount, cartReducer, cartTotalCents, emptyCart, isCart, type Cart, type CartAction } from "../domain/cart";

const storageKey = "shop.cart";

interface CartState {
  cart: Cart;
  totalCents: number;
  itemCount: number;
  dispatch: (action: CartAction) => void;
}

const CartContext = createContext<CartState | null>(null);

function loadCart(): Cart {
  try {
    const parsed: unknown = JSON.parse(localStorage.getItem(storageKey) ?? "null");
    return isCart(parsed) ? parsed : emptyCart;
  } catch {
    return emptyCart;
  }
}

export function CartProvider({ children }: { children: ReactNode }) {
  const [cart, dispatch] = useReducer(cartReducer, undefined, loadCart);

  useEffect(() => {
    try {
      localStorage.setItem(storageKey, JSON.stringify(cart));
    } catch {
      return;
    }
  }, [cart]);

  const value = useMemo(
    () => ({ cart, dispatch, totalCents: cartTotalCents(cart), itemCount: cartItemCount(cart) }),
    [cart],
  );
  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart(): CartState {
  const value = useContext(CartContext);
  if (value === null) {
    throw new Error("useCart must be used inside CartProvider");
  }
  return value;
}
