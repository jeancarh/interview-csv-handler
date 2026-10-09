package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *checkout.Order) error {
	ex := conn(ctx, r.db)
	res, err := ex.ExecContext(ctx,
		`INSERT INTO orders (idempotency_key, request_hash, status, total_cents, card_last4, failure_reason, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		o.IdempotencyKey, o.RequestHash, o.Status, o.TotalCents, o.CardLast4, o.FailureReason,
		formatTime(o.CreatedAt), formatTime(o.UpdatedAt))
	if isUniqueViolation(err) {
		return checkout.ErrDuplicateKey
	}
	if err != nil {
		return err
	}
	if o.ID, err = res.LastInsertId(); err != nil {
		return err
	}

	for _, it := range o.Items {
		if _, err := ex.ExecContext(ctx,
			`INSERT INTO order_items (order_id, product_id, sku, name, unit_price_cents, quantity) VALUES (?, ?, ?, ?, ?, ?)`,
			o.ID, it.ProductID, it.SKU, it.Name, it.UnitPriceCents, it.Quantity); err != nil {
			return err
		}
	}
	return nil
}

func (r *OrderRepository) Get(ctx context.Context, id int64) (checkout.Order, error) {
	return r.load(ctx, `id = ?`, id)
}

func (r *OrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (checkout.Order, error) {
	return r.load(ctx, `idempotency_key = ?`, key)
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, o *checkout.Order) error {
	res, err := conn(ctx, r.db).ExecContext(ctx,
		`UPDATE orders SET status = ?, failure_reason = ?, updated_at = ? WHERE id = ?`,
		o.Status, o.FailureReason, formatTime(o.UpdatedAt), o.ID)
	return expectOneRow(res, err, checkout.ErrNotFound)
}

func (r *OrderRepository) load(ctx context.Context, cond string, arg any) (checkout.Order, error) {
	ex := conn(ctx, r.db)
	var (
		o                checkout.Order
		created, updated string
	)
	err := ex.QueryRowContext(ctx,
		`SELECT id, idempotency_key, request_hash, status, total_cents, card_last4, failure_reason, created_at, updated_at
		 FROM orders WHERE `+cond, arg).
		Scan(&o.ID, &o.IdempotencyKey, &o.RequestHash, &o.Status, &o.TotalCents, &o.CardLast4, &o.FailureReason, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return checkout.Order{}, checkout.ErrNotFound
	}
	if err != nil {
		return checkout.Order{}, err
	}
	if o.CreatedAt, err = parseTime(created); err != nil {
		return checkout.Order{}, err
	}
	if o.UpdatedAt, err = parseTime(updated); err != nil {
		return checkout.Order{}, err
	}

	rows, err := ex.QueryContext(ctx,
		`SELECT product_id, sku, name, unit_price_cents, quantity FROM order_items WHERE order_id = ? ORDER BY product_id`, o.ID)
	if err != nil {
		return checkout.Order{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var it checkout.Item
		if err := rows.Scan(&it.ProductID, &it.SKU, &it.Name, &it.UnitPriceCents, &it.Quantity); err != nil {
			return checkout.Order{}, err
		}
		o.Items = append(o.Items, it)
	}
	return o, rows.Err()
}
