package checkout

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
)

const (
	minKeyLen   = 8
	maxKeyLen   = 100
	maxLines    = 50
	maxQuantity = 100
)

type Line struct {
	ProductID int64
	Quantity  int64
}

type PlaceOrder struct {
	IdempotencyKey string
	Lines          []Line
	CardNumber     string
}

type Service struct {
	products Products
	orders   Orders
	payments PaymentGateway
	tx       Transactor
	now      func() time.Time
}

func NewService(products Products, orders Orders, payments PaymentGateway, tx Transactor, now func() time.Time) *Service {
	return &Service{products: products, orders: orders, payments: payments, tx: tx, now: now}
}

func (s *Service) Get(ctx context.Context, id int64) (Order, error) {
	return s.orders.Get(ctx, id)
}

func (s *Service) PlaceOrder(ctx context.Context, req PlaceOrder) (Order, bool, error) {
	lines, card, err := validate(req)
	if err != nil {
		return Order{}, false, err
	}
	key := strings.TrimSpace(req.IdempotencyKey)
	hash := requestHash(lines, card)

	if existing, found, err := s.replay(ctx, key, hash); found || err != nil {
		return existing, found, err
	}

	order, err := s.reserve(ctx, key, hash, lines, card)
	if errors.Is(err, ErrDuplicateKey) {
		existing, _, err := s.replay(ctx, key, hash)
		return existing, true, err
	}
	if err != nil {
		return Order{}, false, err
	}

	chargeErr := s.payments.Charge(ctx, Charge{
		Reference:   fmt.Sprintf("order-%d", order.ID),
		AmountCents: order.TotalCents,
		CardNumber:  card,
	})
	if chargeErr != nil {
		return s.fail(ctx, order, chargeErr)
	}

	order.MarkPaid(s.now())
	if err := s.orders.UpdateStatus(ctx, &order); err != nil {
		return Order{}, false, err
	}
	return order, false, nil
}

func (s *Service) replay(ctx context.Context, key, hash string) (Order, bool, error) {
	existing, err := s.orders.FindByIdempotencyKey(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return Order{}, false, nil
	}
	if err != nil {
		return Order{}, false, err
	}
	if existing.RequestHash != hash {
		return Order{}, false, ErrIdempotencyConflict
	}
	return existing, true, nil
}

func (s *Service) reserve(ctx context.Context, key, hash string, lines []Line, card string) (Order, error) {
	now := s.now()
	order := Order{
		IdempotencyKey: key,
		RequestHash:    hash,
		Status:         StatusPending,
		CardLast4:      card[len(card)-4:],
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		for _, l := range lines {
			p, err := s.products.Get(ctx, l.ProductID)
			if errors.Is(err, catalog.ErrNotFound) {
				return &ValidationError{Fields: map[string]string{"items": fmt.Sprintf("product %d is not available", l.ProductID)}}
			}
			if err != nil {
				return err
			}
			if err := s.products.ReserveStock(ctx, p.ID, l.Quantity); err != nil {
				if errors.Is(err, catalog.ErrInsufficientStock) {
					return fmt.Errorf("%w for %s (%d left)", err, p.Name, p.Stock)
				}
				return err
			}
			item := Item{ProductID: p.ID, SKU: p.SKU, Name: p.Name, UnitPriceCents: p.PriceCents, Quantity: l.Quantity}
			order.Items = append(order.Items, item)
			order.TotalCents += item.SubtotalCents()
		}
		return s.orders.Create(ctx, &order)
	})
	return order, err
}

func (s *Service) fail(ctx context.Context, order Order, cause error) (Order, bool, error) {
	reason := "payment could not be processed"
	if errors.Is(cause, ErrPaymentDeclined) {
		reason = strings.TrimPrefix(cause.Error(), ErrPaymentDeclined.Error()+": ")
	}
	order.MarkFailed(reason, s.now())

	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		for _, item := range order.Items {
			if err := s.products.ReleaseStock(ctx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		return s.orders.UpdateStatus(ctx, &order)
	})
	if err != nil {
		return Order{}, false, err
	}
	return order, false, nil
}

func validate(req PlaceOrder) ([]Line, string, error) {
	fields := map[string]string{}

	key := strings.TrimSpace(req.IdempotencyKey)
	if len(key) < minKeyLen || len(key) > maxKeyLen {
		fields["idempotencyKey"] = "must be between 8 and 100 characters"
	}

	merged := map[int64]int64{}
	for _, l := range req.Lines {
		if l.Quantity < 1 || l.Quantity > maxQuantity {
			fields["items"] = "each quantity must be between 1 and 100"
			continue
		}
		merged[l.ProductID] += l.Quantity
	}
	switch {
	case len(req.Lines) == 0:
		fields["items"] = "the cart is empty"
	case len(merged) > maxLines:
		fields["items"] = "too many different products in one order"
	}
	lines := make([]Line, 0, len(merged))
	for id, qty := range merged {
		if qty > maxQuantity {
			fields["items"] = "each quantity must be between 1 and 100"
		}
		lines = append(lines, Line{ProductID: id, Quantity: qty})
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i].ProductID < lines[j].ProductID })

	card := strings.NewReplacer(" ", "", "-", "").Replace(req.CardNumber)
	if len(card) < 12 || len(card) > 19 || strings.Trim(card, "0123456789") != "" {
		fields["cardNumber"] = "must be 12 to 19 digits"
	}

	if len(fields) > 0 {
		return nil, "", &ValidationError{Fields: fields}
	}
	return lines, card, nil
}

func requestHash(lines []Line, card string) string {
	h := sha256.New()
	for _, l := range lines {
		fmt.Fprintf(h, "%d:%d;", l.ProductID, l.Quantity)
	}
	fmt.Fprintf(h, "card:%s", card[len(card)-4:])
	return hex.EncodeToString(h.Sum(nil))
}
