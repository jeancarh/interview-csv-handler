import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createPortContext } from "../../../shared/di/createPortContext";
import type { ProductDraft } from "../domain/product";
import type { ProductRepository, ProductSearch } from "../domain/productRepository";

export const [ProductRepositoryProvider, useProductRepository] = createPortContext<ProductRepository>("ProductRepository");

export const productKeys = {
  all: ["products"] as const,
  search: (query: ProductSearch) => ["products", "search", query] as const,
  detail: (id: number) => ["products", "detail", id] as const,
  categories: ["categories"] as const,
};

export function useProductSearch(query: ProductSearch) {
  const repo = useProductRepository();
  return useQuery({
    queryKey: productKeys.search(query),
    queryFn: () => repo.search(query),
    placeholderData: keepPreviousData,
    refetchInterval: 30_000,
  });
}

export function useProduct(id: number | null) {
  const repo = useProductRepository();
  return useQuery({
    queryKey: productKeys.detail(id ?? 0),
    queryFn: () => repo.get(id!),
    enabled: id !== null,
  });
}

export function useCategories() {
  const repo = useProductRepository();
  return useQuery({ queryKey: productKeys.categories, queryFn: () => repo.categories(), staleTime: 60_000 });
}

export function useSaveProduct(id: number | null) {
  const repo = useProductRepository();
  const client = useQueryClient();
  return useMutation({
    mutationFn: (draft: ProductDraft) => (id === null ? repo.create(draft) : repo.update(id, draft)),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: productKeys.all });
      client.invalidateQueries({ queryKey: productKeys.categories });
    },
  });
}

export function useDeleteProduct() {
  const repo = useProductRepository();
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => repo.remove(id),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: productKeys.all });
      client.invalidateQueries({ queryKey: productKeys.categories });
    },
  });
}
