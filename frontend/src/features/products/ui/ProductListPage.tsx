import { useState } from "react";
import { Link } from "react-router";
import { Alert, Pagination, errorMessage } from "../../../shared/ui/components";
import { useDebouncedValue } from "../../../shared/ui/useDebouncedValue";
import { useCart } from "../../cart/application/cart";
import { useCategories, useDeleteProduct, useProductSearch } from "../application/products";
import type { Product } from "../domain/product";
import { ProductTable } from "./ProductTable";
import { SearchFilters } from "./SearchFilters";

const pageSize = 20;

export function ProductListPage() {
  const [text, setText] = useState("");
  const [category, setCategory] = useState("");
  const [inStock, setInStock] = useState(false);
  const [page, setPage] = useState(1);
  const [notice, setNotice] = useState<string | null>(null);
  const q = useDebouncedValue(text, 250);

  const search = useProductSearch({ q, category, inStock, page, pageSize });
  const categories = useCategories();
  const remove = useDeleteProduct();
  const { dispatch } = useCart();

  function resetPage<T>(set: (v: T) => void) {
    return (value: T) => {
      set(value);
      setPage(1);
    };
  }

  function addToCart(product: Product) {
    dispatch({ type: "add", product });
    setNotice(`${product.name} added to the cart`);
  }

  function confirmDelete(product: Product) {
    if (window.confirm(`Delete "${product.name}"? Past orders keep their copy of it.`)) {
      remove.mutate(product.id, { onSuccess: () => setNotice(`${product.name} deleted`) });
    }
  }

  return (
    <section>
      <header className="page-header">
        <h1>Products</h1>
        <div className="actions">
          <Link to="/import" className="button">
            Import CSV
          </Link>
          <Link to="/products/new" className="button primary">
            New product
          </Link>
        </div>
      </header>

      <SearchFilters
        text={text}
        category={category}
        inStock={inStock}
        categories={categories.data ?? []}
        onTextChange={resetPage(setText)}
        onCategoryChange={resetPage(setCategory)}
        onInStockChange={resetPage(setInStock)}
      />

      {notice && <Alert tone="success">{notice}</Alert>}
      {remove.error && <Alert>{errorMessage(remove.error)}</Alert>}
      {search.error && <Alert>{errorMessage(search.error)}</Alert>}

      {search.data && (
        <>
          <p className="muted small">{search.data.total} products</p>
          <ProductTable
            products={search.data.items}
            onAddToCart={addToCart}
            onDelete={confirmDelete}
            deletingId={remove.isPending ? (remove.variables ?? null) : null}
          />
          <Pagination page={page} pageSize={pageSize} total={search.data.total} onChange={setPage} />
        </>
      )}
      {search.isPending && <p className="muted">Loading products...</p>}
    </section>
  );
}
