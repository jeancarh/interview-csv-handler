import { formatCents } from "../../../shared/money/money";
import type { CartLine } from "../domain/cart";

interface CartLinesProps {
  lines: CartLine[];
  totalCents: number;
  onQuantityChange: (productId: number, quantity: number) => void;
  onRemove: (productId: number) => void;
}

export function CartLines({ lines, totalCents, onQuantityChange, onRemove }: CartLinesProps) {
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Product</th>
            <th className="num">Price</th>
            <th className="num">Quantity</th>
            <th className="num">Subtotal</th>
            <th aria-label="Actions" />
          </tr>
        </thead>
        <tbody>
          {lines.map((l) => (
            <tr key={l.productId}>
              <td>
                <div className="product-name">{l.name}</div>
                <div className="muted small mono">{l.sku}</div>
              </td>
              <td className="num">{formatCents(l.unitPriceCents)}</td>
              <td className="num">
                <input
                  className="qty"
                  type="number"
                  min={0}
                  max={Math.min(l.available, 100)}
                  value={l.quantity}
                  aria-label={`Quantity of ${l.name}`}
                  onChange={(e) => onQuantityChange(l.productId, Number(e.target.value))}
                />
              </td>
              <td className="num">{formatCents(l.unitPriceCents * l.quantity)}</td>
              <td className="num">
                <button type="button" className="danger" onClick={() => onRemove(l.productId)}>
                  Remove
                </button>
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr>
            <th colSpan={3}>Total</th>
            <th className="num">{formatCents(totalCents)}</th>
            <th />
          </tr>
        </tfoot>
      </table>
    </div>
  );
}
