package checkout

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
	StatusFailed  Status = "failed"
)

var (
	ErrNotFound            = errors.New("order not found")
	ErrIdempotencyConflict = errors.New("idempotency key was already used with a different request")
	ErrDuplicateKey        = errors.New("idempotency key already exists")
	ErrPaymentDeclined     = errors.New("payment declined")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+": "+e.Fields[k])
	}
	return "invalid order: " + strings.Join(parts, "; ")
}

type Item struct {
	ProductID      int64
	SKU            string
	Name           string
	UnitPriceCents int64
	Quantity       int64
}

func (i Item) SubtotalCents() int64 {
	return i.UnitPriceCents * i.Quantity
}

type Order struct {
	ID             int64
	IdempotencyKey string
	RequestHash    string
	Status         Status
	Items          []Item
	TotalCents     int64
	CardLast4      string
	FailureReason  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (o *Order) MarkPaid(now time.Time) {
	o.Status = StatusPaid
	o.UpdatedAt = now
}

func (o *Order) MarkFailed(reason string, now time.Time) {
	o.Status = StatusFailed
	o.FailureReason = reason
	o.UpdatedAt = now
}
