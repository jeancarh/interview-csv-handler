import { NavLink, Outlet } from "react-router";
import { useCart } from "../features/cart/application/cart";

export function Layout() {
  const { itemCount } = useCart();

  return (
    <>
      <header className="topbar">
        <nav>
          <NavLink to="/" end className="brand">
            Shop
          </NavLink>
          <NavLink to="/" end>
            Products
          </NavLink>
          <NavLink to="/import">Import</NavLink>
          <NavLink to="/cart">Cart{itemCount > 0 && <span className="count">{itemCount}</span>}</NavLink>
          <a href="/docs" target="_blank" rel="noreferrer">
            API docs
          </a>
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
    </>
  );
}
