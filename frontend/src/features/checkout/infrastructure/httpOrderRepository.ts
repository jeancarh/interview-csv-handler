import type { components } from "../../../shared/api/schema";
import { ApiError, request } from "../../../shared/http/client";
import { ValidationFailure } from "../../../shared/validation";
import type { Order, OrderRepository } from "../domain/order";

type OrderDto = components["schemas"]["Order"];
type OrderInputDto = components["schemas"]["OrderInput"];

function toOrder(dto: OrderDto): Order {
  return {
    id: dto.id,
    status: dto.status,
    items: dto.items,
    totalCents: dto.totalCents,
    cardLast4: dto.cardLast4,
    failureReason: dto.failureReason ?? "",
    createdAt: dto.createdAt,
  };
}

export const httpOrderRepository: OrderRepository = {
  async place(checkout, idempotencyKey) {
    const body: OrderInputDto = { items: checkout.lines, cardNumber: checkout.cardNumber };
    try {
      const { data, headers } = await request<OrderDto>("/api/v1/orders", {
        method: "POST",
        headers: { "Idempotency-Key": idempotencyKey },
        body: JSON.stringify(body),
      });
      return { order: toOrder(data), replayed: headers.get("Idempotent-Replayed") === "true" };
    } catch (error) {
      if (error instanceof ApiError && error.status === 422) {
        throw new ValidationFailure(error.fields, error.message);
      }
      throw error;
    }
  },

  async get(id) {
    const { data } = await request<OrderDto>(`/api/v1/orders/${id}`);
    return toOrder(data);
  },
};
