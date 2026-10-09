import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRef } from "react";
import { createPortContext } from "../../../shared/di/createPortContext";
import { fingerprint, type CheckoutRequest, type OrderRepository } from "../domain/order";

export const [OrderRepositoryProvider, useOrderRepository] = createPortContext<OrderRepository>("OrderRepository");

export function usePlaceOrder() {
  const repo = useOrderRepository();
  const client = useQueryClient();
  const attempt = useRef<{ fingerprint: string; key: string } | null>(null);

  return useMutation({
    mutationFn: (request: CheckoutRequest) => {
      const current = fingerprint(request);
      if (attempt.current?.fingerprint !== current) {
        attempt.current = { fingerprint: current, key: crypto.randomUUID() };
      }
      return repo.place(request, attempt.current.key);
    },
    onSuccess: ({ order }) => {
      attempt.current = null;
      client.setQueryData(["orders", order.id], order);
      client.invalidateQueries({ queryKey: ["products"] });
    },
  });
}

export function useOrder(id: number) {
  const repo = useOrderRepository();
  return useQuery({ queryKey: ["orders", id], queryFn: () => repo.get(id) });
}
