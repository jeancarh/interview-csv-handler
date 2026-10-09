package catalog

import (
	"context"
	"time"
)

type memRepo struct {
	nextID   int64
	products map[int64]Product
}

func newMemRepo() *memRepo {
	return &memRepo{products: map[int64]Product{}}
}

func (r *memRepo) Create(_ context.Context, p *Product) error {
	if _, err := r.FindBySKU(context.Background(), p.SKU); err == nil {
		return ErrDuplicateSKU
	}
	r.nextID++
	p.ID = r.nextID
	r.products[p.ID] = *p
	return nil
}

func (r *memRepo) Update(_ context.Context, p *Product) error {
	if _, ok := r.products[p.ID]; !ok {
		return ErrNotFound
	}
	r.products[p.ID] = *p
	return nil
}

func (r *memRepo) Get(_ context.Context, id int64) (Product, error) {
	p, ok := r.products[id]
	if !ok {
		return Product{}, ErrNotFound
	}
	return p, nil
}

func (r *memRepo) FindBySKU(_ context.Context, sku string) (Product, error) {
	for _, p := range r.products {
		if p.SKU == sku {
			return p, nil
		}
	}
	return Product{}, ErrNotFound
}

func (r *memRepo) Delete(_ context.Context, id int64, _ time.Time) error {
	if _, ok := r.products[id]; !ok {
		return ErrNotFound
	}
	delete(r.products, id)
	return nil
}

func (r *memRepo) Search(_ context.Context, q SearchQuery) (Page, error) {
	items := make([]Product, 0, len(r.products))
	for _, p := range r.products {
		items = append(items, p)
	}
	return Page{Items: items, Total: len(items), Page: q.Page, PageSize: q.PageSize}, nil
}

func (r *memRepo) Categories(context.Context) ([]string, error) {
	return nil, nil
}

type noTx struct{}

func (noTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func fixedClock() time.Time {
	return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
}
