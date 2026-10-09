import type { Product, ProductDraft } from "./product";

export interface ProductSearch {
  q: string;
  category: string;
  inStock: boolean;
  page: number;
  pageSize: number;
}

export interface ProductPage {
  items: Product[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ProductRepository {
  search(query: ProductSearch): Promise<ProductPage>;
  get(id: number): Promise<Product>;
  create(draft: ProductDraft): Promise<Product>;
  update(id: number, draft: ProductDraft): Promise<Product>;
  remove(id: number): Promise<void>;
  categories(): Promise<string[]>;
}
