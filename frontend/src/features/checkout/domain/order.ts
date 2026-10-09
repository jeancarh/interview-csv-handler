export type OrderStatus = "pending" | "paid" | "failed";

export interface OrderItem {
  productId: number;
  sku: string;
  name: string;
  unitPriceCents: number;
  quantity: number;
  subtotalCents: number;
}

export interface Order {
  id: number;
  status: OrderStatus;
  items: OrderItem[];
  totalCents: number;
  cardLast4: string;
  failureReason: string;
  createdAt: string;
}

export interface CheckoutRequest {
  lines: { productId: number; quantity: number }[];
  cardNumber: string;
}

export interface PlacedOrder {
  order: Order;
  replayed: boolean;
}

export interface OrderRepository {
  place(request: CheckoutRequest, idempotencyKey: string): Promise<PlacedOrder>;
  get(id: number): Promise<Order>;
}

export function normalizeCardNumber(input: string): string {
  return input.replace(/[\s-]/g, "");
}

export function validateCardNumber(input: string): string | null {
  const digits = normalizeCardNumber(input);
  if (!/^\d{12,19}$/.test(digits)) {
    return "Enter a card number with 12 to 19 digits";
  }
  return null;
}

export function fingerprint(request: CheckoutRequest): string {
  const lines = [...request.lines]
    .sort((a, b) => a.productId - b.productId)
    .map((l) => `${l.productId}x${l.quantity}`)
    .join(",");
  return `${lines}|${normalizeCardNumber(request.cardNumber)}`;
}
