interface SearchFiltersProps {
  text: string;
  category: string;
  inStock: boolean;
  categories: string[];
  onTextChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  onInStockChange: (value: boolean) => void;
}

export function SearchFilters(props: SearchFiltersProps) {
  return (
    <div className="filters" role="search">
      <input
        type="search"
        placeholder="Search by name, SKU or description"
        aria-label="Search products"
        value={props.text}
        onChange={(e) => props.onTextChange(e.target.value)}
      />
      <select aria-label="Category" value={props.category} onChange={(e) => props.onCategoryChange(e.target.value)}>
        <option value="">All categories</option>
        {props.categories.map((c) => (
          <option key={c} value={c}>
            {c}
          </option>
        ))}
      </select>
      <label className="checkbox">
        <input type="checkbox" checked={props.inStock} onChange={(e) => props.onInStockChange(e.target.checked)} />
        In stock only
      </label>
    </div>
  );
}
