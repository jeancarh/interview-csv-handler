import type { components } from "../../../shared/api/schema";
import { ApiError, request } from "../../../shared/http/client";
import { ValidationFailure } from "../../../shared/validation";
import { toFormErrors, type Product, type ProductDraft } from "../domain/product";
import type { ProductPage, ProductRepository, ProductSearch } from "../domain/productRepository";

type ProductDto = components["schemas"]["Product"];
type ProductInputDto = components["schemas"]["ProductInput"];
type ProductPageDto = components["schemas"]["ProductPage"];

const base = "/api/v1/products";

function toProduct(dto: ProductDto): Product {
  return {
    id: dto.id,
    sku: dto.sku,
    name: dto.name,
    description: dto.description ?? "",
    category: dto.category,
    priceCents: dto.priceCents,
    stock: dto.stock,
    weightGrams: dto.weightGrams ?? null,
  };
}

function toInput(draft: ProductDraft): ProductInputDto {
  return { ...draft };
}

async function translateErrors<T>(call: Promise<T>): Promise<T> {
  try {
    return await call;
  } catch (error) {
    if (error instanceof ApiError && error.status === 422) {
      throw new ValidationFailure(toFormErrors(error.fields), error.message);
    }
    if (error instanceof ApiError && error.status === 409) {
      throw new ValidationFailure({ sku: "Another product already uses this SKU" }, error.message);
    }
    throw error;
  }
}

export const httpProductRepository: ProductRepository = {
  async search(query: ProductSearch): Promise<ProductPage> {
    const params = new URLSearchParams({ page: String(query.page), pageSize: String(query.pageSize) });
    if (query.q.trim()) params.set("q", query.q.trim());
    if (query.category) params.set("category", query.category);
    if (query.inStock) params.set("inStock", "true");

    const { data } = await request<ProductPageDto>(`${base}?${params}`);
    return { ...data, items: data.items.map(toProduct) };
  },

  async get(id) {
    const { data } = await request<ProductDto>(`${base}/${id}`);
    return toProduct(data);
  },

  async create(draft) {
    const { data } = await translateErrors(
      request<ProductDto>(base, { method: "POST", body: JSON.stringify(toInput(draft)) }),
    );
    return toProduct(data);
  },

  async update(id, draft) {
    const { data } = await translateErrors(
      request<ProductDto>(`${base}/${id}`, { method: "PUT", body: JSON.stringify(toInput(draft)) }),
    );
    return toProduct(data);
  },

  async remove(id) {
    await request<void>(`${base}/${id}`, { method: "DELETE" });
  },

  async categories() {
    const { data } = await request<{ items: string[] }>("/api/v1/categories");
    return data.items;
  },
};
