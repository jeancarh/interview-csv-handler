package checkout

import (
	"context"

	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
)

type Products interface {
	Get(ctx context.Context, id int64) (catalog.Product, error)
	ReserveStock(ctx context.Context, id, qty int64) error
	ReleaseStock(ctx context.Context, id, qty int64) error
}

type Orders interface {
	Create(ctx context.Context, o *Order) error
	Get(ctx context.Context, id int64) (Order, error)
	FindByIdempotencyKey(ctx context.Context, key string) (Order, error)
	UpdateStatus(ctx context.Context, o *Order) error
}

type PaymentGateway interface {
	Charge(ctx context.Context, c Charge) error
}

type Charge struct {
	Reference   string
	AmountCents int64
	CardNumber  string
}

type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
