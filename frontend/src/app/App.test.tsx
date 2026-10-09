import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { describe, expect, it, vi } from "vitest";
import type { Order, OrderRepository } from "../features/checkout/domain/order";
import type { ImportReport } from "../features/import/domain/importReport";
import type { Product } from "../features/products/domain/product";
import type { ProductRepository } from "../features/products/domain/productRepository";
import { ValidationFailure } from "../shared/validation";
import { createQueryClient, Providers, routes, type Adapters } from "./App";

const catalog: Product[] = [
  { id: 1, sku: "RS-001", name: "Running Shoes", description: "Daily trainer", category: "Footwear", priceCents: 8999, stock: 3, weightGrams: 350 },
  { id: 2, sku: "XS-001", name: "<script>alert('xss')</script>", description: "", category: "Electronics", priceCents: 1999, stock: 100, weightGrams: 100 },
  { id: 3, sku: "VC-001", name: "Vintage Clock", description: "", category: "Home & Office", priceCents: 29999, stock: 0, weightGrams: 3500 },
];

function memoryProducts(items: Product[] = catalog): ProductRepository {
  return {
    search: async (q) => {
      const found = items.filter((p) => p.name.toLowerCase().includes(q.q.toLowerCase()));
      return { items: found, total: found.length, page: q.page, pageSize: q.pageSize };
    },
    get: async (id) => items.find((p) => p.id === id)!,
    create: async (draft) => ({ id: 99, ...draft }),
    update: async (id, draft) => ({ id, ...draft }),
    remove: async () => undefined,
    categories: async () => [...new Set(items.map((p) => p.category))],
  };
}

const paidOrder: Order = {
  id: 7,
  status: "paid",
  items: [{ productId: 1, sku: "RS-001", name: "Running Shoes", unitPriceCents: 8999, quantity: 1, subtotalCents: 8999 }],
  totalCents: 8999,
  cardLast4: "4242",
  failureReason: "",
  createdAt: "2026-10-09T12:00:00Z",
};

function renderApp(path: string, overrides: Partial<Adapters> = {}) {
  const adapters: Adapters = {
    products: memoryProducts(),
    imports: { upload: vi.fn() },
    orders: { place: vi.fn(), get: async () => paidOrder },
    ...overrides,
  };
  const router = createMemoryRouter(routes, { initialEntries: [path] });
  render(
    <Providers adapters={adapters} client={createQueryClient()}>
      <RouterProvider router={router} />
    </Providers>,
  );
  return { router, adapters };
}

describe("product list", () => {
  it("renders hostile names as plain text", async () => {
    renderApp("/");
    expect(await screen.findByText("<script>alert('xss')</script>")).toBeInTheDocument();
    expect(document.querySelector("main script")).toBeNull();
  });

  it("disables add to cart for sold out products", async () => {
    renderApp("/");
    const row = (await screen.findByText("Vintage Clock")).closest("tr")!;
    expect(within(row).getByRole("button", { name: "Add to cart" })).toBeDisabled();
  });

  it("adds products to the cart", async () => {
    const user = userEvent.setup();
    renderApp("/");
    const row = (await screen.findByText("Running Shoes")).closest("tr")!;
    await user.click(within(row).getByRole("button", { name: "Add to cart" }));
    expect(screen.getByRole("link", { name: /Cart\s*1/ })).toBeInTheDocument();
  });
});

describe("product form", () => {
  it("validates on the client before calling the API", async () => {
    const user = userEvent.setup();
    const create = vi.fn();
    renderApp("/products/new", { products: { ...memoryProducts(), create } });

    await user.click(await screen.findByRole("button", { name: "Create product" }));
    expect(screen.getByText("Name is required")).toBeInTheDocument();
    expect(create).not.toHaveBeenCalled();
  });

  it("shows field errors coming from the API", async () => {
    const user = userEvent.setup();
    const create = vi.fn().mockRejectedValue(new ValidationFailure({ price: "can't be more than 1,000,000.00" }));
    renderApp("/products/new", { products: { ...memoryProducts(), create } });

    await user.type(await screen.findByLabelText("Name"), "Desk");
    await user.type(screen.getByLabelText("SKU"), "DK-1");
    await user.type(screen.getByLabelText("Category"), "Home");
    await user.type(screen.getByLabelText("Price (USD)"), "9999999");
    await user.click(screen.getByRole("button", { name: "Create product" }));

    expect(await screen.findByText("can't be more than 1,000,000.00")).toBeInTheDocument();
  });
});

describe("checkout", () => {
  it("reuses the idempotency key when retrying after a network error", async () => {
    const user = userEvent.setup();
    const place = vi
      .fn<OrderRepository["place"]>()
      .mockRejectedValueOnce(new Error("Could not reach the server"))
      .mockResolvedValueOnce({ order: paidOrder, replayed: true });
    localStorage.setItem(
      "shop.cart",
      JSON.stringify({ lines: [{ productId: 1, sku: "RS-001", name: "Running Shoes", unitPriceCents: 8999, quantity: 1, available: 3 }] }),
    );
    const { router } = renderApp("/cart", { orders: { place, get: async () => paidOrder } });

    await user.type(await screen.findByLabelText("Card number"), "4242 4242 4242 4242");
    await user.click(screen.getByRole("button", { name: /Pay/ }));
    expect(await screen.findByText("Could not reach the server")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Pay/ }));
    await waitFor(() => expect(router.state.location.pathname).toBe("/orders/7"));

    const firstKey = place.mock.calls[0]?.[1];
    expect(firstKey).toBeTruthy();
    expect(place.mock.calls[1]?.[1]).toBe(firstKey);
  });

  it("blocks obviously wrong card numbers", async () => {
    const user = userEvent.setup();
    const place = vi.fn();
    localStorage.setItem(
      "shop.cart",
      JSON.stringify({ lines: [{ productId: 1, sku: "RS-001", name: "Running Shoes", unitPriceCents: 8999, quantity: 1, available: 3 }] }),
    );
    renderApp("/cart", { orders: { place, get: async () => paidOrder } });

    await user.type(await screen.findByLabelText("Card number"), "1234");
    await user.click(screen.getByRole("button", { name: /Pay/ }));
    expect(screen.getByText("Enter a card number with 12 to 19 digits")).toBeInTheDocument();
    expect(place).not.toHaveBeenCalled();
  });
});

describe("import", () => {
  it("shows the row report and filters by status", async () => {
    const user = userEvent.setup();
    const report: ImportReport = {
      summary: { total: 2, created: 1, updated: 0, unchanged: 0, skipped: 0, rejected: 1 },
      rows: [
        { line: 2, sku: "RS-001", name: "Running Shoes", status: "created", messages: [] },
        { line: 7, sku: "YM-015", name: "Yoga Mat", status: "rejected", messages: ['price "free" is not a valid amount'] },
      ],
    };
    renderApp("/import", { imports: { upload: vi.fn().mockResolvedValue(report) } });

    const file = new File(["name,sku\n"], "products.csv", { type: "text/csv" });
    await user.upload(await screen.findByLabelText("CSV file"), file);
    await user.click(screen.getByRole("button", { name: "Import" }));

    expect(await screen.findByText('price "free" is not a valid amount')).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /rejected/ }));
    expect(screen.queryByText("Running Shoes")).not.toBeInTheDocument();
  });
});
