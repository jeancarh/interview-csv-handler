import { Link, useParams } from "react-router";
import { formatCents } from "../../../shared/money/money";
import { Alert, errorMessage } from "../../../shared/ui/components";
import { useOrder } from "../application/checkout";
import type { Order } from "../domain/order";

export function OrderPage() {
  const id = Number(useParams().id);
  const order = useOrder(id);

  if (order.isPending) {
    return <p className="muted">Loading order...</p>;
  }
  if (order.error) {
    return <Alert>{errorMessage(order.error)}</Alert>;
  }
  return <OrderSummary order={order.data} />;
}

export function OrderSummary({ order }: { order: Order }) {
  return (
    <section className="narrow">
      <h1>Order #{order.id}</h1>
      {order.status === "paid" && <Alert tone="success">Payment approved. Thanks for your order.</Alert>}
      {order.status === "failed" && (
        <Alert>
          Payment failed: {order.failureReason}. Nothing was charged and your cart is still there.{" "}
          <Link to="/cart">Back to the cart</Link>
        </Alert>
      )}
      {order.status === "pending" && <Alert tone="info">Payment is still being processed.</Alert>}

      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Product</th>
              <th className="num">Price</th>
              <th className="num">Qty</th>
              <th className="num">Subtotal</th>
            </tr>
          </thead>
          <tbody>
            {order.items.map((it) => (
              <tr key={it.productId}>
                <td>{it.name}</td>
                <td className="num">{formatCents(it.unitPriceCents)}</td>
                <td className="num">{it.quantity}</td>
                <td className="num">{formatCents(it.subtotalCents)}</td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr>
              <th colSpan={3}>Total</th>
              <th className="num">{formatCents(order.totalCents)}</th>
            </tr>
          </tfoot>
        </table>
      </div>
      <p className="muted small">Card ending in {order.cardLast4}</p>
      <Link to="/" className="button">
        Continue shopping
      </Link>
    </section>
  );
}
