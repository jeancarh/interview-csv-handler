import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { Alert, errorMessage } from "../../../shared/ui/components";
import { ValidationFailure } from "../../../shared/validation";
import { useCart } from "../../cart/application/cart";
import { CartLines } from "../../cart/ui/CartLines";
import { usePlaceOrder } from "../application/checkout";
import { normalizeCardNumber, validateCardNumber } from "../domain/order";
import { CheckoutForm } from "./CheckoutForm";

export function CartPage() {
  const { cart, totalCents, dispatch } = useCart();
  const placeOrder = usePlaceOrder();
  const navigate = useNavigate();
  const [cardError, setCardError] = useState<string | undefined>();

  function pay(cardNumber: string) {
    const problem = validateCardNumber(cardNumber);
    setCardError(problem ?? undefined);
    if (problem) {
      return;
    }
    const lines = cart.lines.map((l) => ({ productId: l.productId, quantity: l.quantity }));
    placeOrder.mutate(
      { lines, cardNumber: normalizeCardNumber(cardNumber) },
      {
        onSuccess: ({ order }) => {
          if (order.status === "paid") {
            dispatch({ type: "clear" });
          }
          navigate(`/orders/${order.id}`);
        },
        onError: (error) => {
          if (error instanceof ValidationFailure && error.fields.cardNumber) {
            setCardError(error.fields.cardNumber);
          }
        },
      },
    );
  }

  if (cart.lines.length === 0) {
    return (
      <section>
        <h1>Cart</h1>
        <p className="empty">
          Your cart is empty. <Link to="/">Browse products</Link>
        </p>
      </section>
    );
  }

  return (
    <section>
      <h1>Cart</h1>
      {placeOrder.error && <Alert>{errorMessage(placeOrder.error)}</Alert>}
      <CartLines
        lines={cart.lines}
        totalCents={totalCents}
        onQuantityChange={(productId, quantity) => dispatch({ type: "setQuantity", productId, quantity })}
        onRemove={(productId) => dispatch({ type: "remove", productId })}
      />
      <CheckoutForm totalCents={totalCents} submitting={placeOrder.isPending} cardError={cardError} onPay={pay} />
    </section>
  );
}
