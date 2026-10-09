package catalog

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Get(ctx context.Context, id int64) (Product, error)
	FindBySKU(ctx context.Context, sku string) (Product, error)
	Delete(ctx context.Context, id int64, at time.Time) error
	Search(ctx context.Context, q SearchQuery) (Page, error)
	Categories(ctx context.Context) ([]string, error)
}

type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type SearchQuery struct {
	Text          string
	Category      string
	MinPriceCents *int64
	MaxPriceCents *int64
	InStockOnly   bool
	Page          int
	PageSize      int
}

type Page struct {
	Items    []Product
	Total    int
	Page     int
	PageSize int
}
