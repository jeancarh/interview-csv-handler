import { Link } from "react-router";
import { formatCents } from "../../../shared/money/money";
import type { Product } from "../domain/product";

interface ProductTableProps {
  products: Product[];
  onAddToCart: (product: Product) => void;
  onDelete: (product: Product) => void;
  deletingId: number | null;
}

export function ProductTable({ products, onAddToCart, onDelete, deletingId }: ProductTableProps) {
  if (products.length === 0) {
    return <p className="empty">No products match your search.</p>;
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>SKU</th>
            <th>Category</th>
            <th className="num">Price</th>
            <th className="num">Stock</th>
            <th aria-label="Actions" />
          </tr>
        </thead>
        <tbody>
          {products.map((p) => (
            <tr key={p.id}>
              <td>
                <div className="product-name">{p.name}</div>
                {p.description && <div className="muted small">{p.description}</div>}
              </td>
              <td className="mono">{p.sku}</td>
              <td>{p.category}</td>
              <td className="num">{formatCents(p.priceCents)}</td>
              <td className="num">{p.stock > 0 ? p.stock : <span className="badge badge-rejected">Sold out</span>}</td>
              <td>
                <div className="actions">
                  <button type="button" className="primary" onClick={() => onAddToCart(p)} disabled={p.stock < 1}>
                    Add to cart
                  </button>
                  <Link to={`/products/${p.id}/edit`} className="button">
                    Edit
                  </Link>
                  <button type="button" className="danger" onClick={() => onDelete(p)} disabled={deletingId === p.id}>
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
