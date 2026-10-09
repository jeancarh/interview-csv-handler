import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { createBrowserRouter, RouterProvider } from "react-router";
import { CartProvider } from "../features/cart/application/cart";
import { OrderRepositoryProvider } from "../features/checkout/application/checkout";
import type { OrderRepository } from "../features/checkout/domain/order";
import { httpOrderRepository } from "../features/checkout/infrastructure/httpOrderRepository";
import { CartPage } from "../features/checkout/ui/CartPage";
import { OrderPage } from "../features/checkout/ui/OrderPage";
import { ImportRepositoryProvider } from "../features/import/application/importProducts";
import type { ImportRepository } from "../features/import/domain/importReport";
import { httpImportRepository } from "../features/import/infrastructure/httpImportRepository";
import { ImportPage } from "../features/import/ui/ImportPage";
import { ProductRepositoryProvider } from "../features/products/application/products";
import type { ProductRepository } from "../features/products/domain/productRepository";
import { httpProductRepository } from "../features/products/infrastructure/httpProductRepository";
import { ProductFormPage } from "../features/products/ui/ProductFormPage";
import { ProductListPage } from "../features/products/ui/ProductListPage";
import { Layout } from "./Layout";

export interface Adapters {
  products: ProductRepository;
  imports: ImportRepository;
  orders: OrderRepository;
}

export const httpAdapters: Adapters = {
  products: httpProductRepository,
  imports: httpImportRepository,
  orders: httpOrderRepository,
};

export const routes = [
  {
    element: <Layout />,
    children: [
      { path: "/", element: <ProductListPage /> },
      { path: "/products/new", element: <ProductFormPage /> },
      { path: "/products/:id/edit", element: <ProductFormPage /> },
      { path: "/import", element: <ImportPage /> },
      { path: "/cart", element: <CartPage /> },
      { path: "/orders/:id", element: <OrderPage /> },
      { path: "*", element: <p className="empty">Page not found.</p> },
    ],
  },
];

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: 1, staleTime: 5_000 },
      mutations: { retry: false },
    },
  });
}

export function Providers({ adapters, client, children }: { adapters: Adapters; client: QueryClient; children: ReactNode }) {
  return (
    <QueryClientProvider client={client}>
      <ProductRepositoryProvider value={adapters.products}>
        <ImportRepositoryProvider value={adapters.imports}>
          <OrderRepositoryProvider value={adapters.orders}>
            <CartProvider>{children}</CartProvider>
          </OrderRepositoryProvider>
        </ImportRepositoryProvider>
      </ProductRepositoryProvider>
    </QueryClientProvider>
  );
}

const router = createBrowserRouter(routes);
const client = createQueryClient();

export function App() {
  return (
    <Providers adapters={httpAdapters} client={client}>
      <RouterProvider router={router} />
    </Providers>
  );
}
