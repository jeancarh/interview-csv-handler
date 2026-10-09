package checkout

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
)

type fakeProducts struct {
	items map[int64]catalog.Product
}

func (f *fakeProducts) Get(_ context.Context, id int64) (catalog.Product, error) {
	p, ok := f.items[id]
	if !ok {
		return catalog.Product{}, catalog.ErrNotFound
	}
	return p, nil
}

func (f *fakeProducts) ReserveStock(_ context.Context, id, qty int64) error {
	p := f.items[id]
	if p.Stock < qty {
		return catalog.ErrInsufficientStock
	}
	p.Stock -= qty
	f.items[id] = p
	return nil
}

func (f *fakeProducts) ReleaseStock(_ context.Context, id, qty int64) error {
	p := f.items[id]
	p.Stock += qty
	f.items[id] = p
	return nil
}

type fakeOrders struct {
	nextID int64
	byID   map[int64]Order
}

func (f *fakeOrders) Create(_ context.Context, o *Order) error {
	for _, existing := range f.byID {
		if existing.IdempotencyKey == o.IdempotencyKey {
			return ErrDuplicateKey
		}
	}
	f.nextID++
	o.ID = f.nextID
	f.byID[o.ID] = *o
	return nil
}

func (f *fakeOrders) Get(_ context.Context, id int64) (Order, error) {
	o, ok := f.byID[id]
	if !ok {
		return Order{}, ErrNotFound
	}
	return o, nil
}

func (f *fakeOrders) FindByIdempotencyKey(_ context.Context, key string) (Order, error) {
	for _, o := range f.byID {
		if o.IdempotencyKey == key {
			return o, nil
		}
	}
	return Order{}, ErrNotFound
}

func (f *fakeOrders) UpdateStatus(_ context.Context, o *Order) error {
	f.byID[o.ID] = *o
	return nil
}

type fakeGateway struct {
	calls   int
	decline bool
}

func (g *fakeGateway) Charge(context.Context, Charge) error {
	g.calls++
	if g.decline {
		return fmt.Errorf("%w: insufficient funds", ErrPaymentDeclined)
	}
	return nil
}

type noTx struct{}

func (noTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fixture struct {
	svc      *Service
	products *fakeProducts
	orders   *fakeOrders
	gateway  *fakeGateway
}

func newFixture() fixture {
	products := &fakeProducts{items: map[int64]catalog.Product{
		1: {ID: 1, SKU: "RS-001", Name: "Running Shoes", PriceCents: 8999, Stock: 10},
		2: {ID: 2, SKU: "YB-016", Name: "Yoga Block", PriceCents: 1299, Stock: 1},
	}}
	orders := &fakeOrders{byID: map[int64]Order{}}
	gateway := &fakeGateway{}
	now := func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }
	return fixture{
		svc:      NewService(products, orders, gateway, noTx{}, now),
		products: products,
		orders:   orders,
		gateway:  gateway,
	}
}

func request(key string, lines ...Line) PlaceOrder {
	return PlaceOrder{IdempotencyKey: key, Lines: lines, CardNumber: "4242 4242 4242 4242"}
}

func TestPlaceOrderPaid(t *testing.T) {
	f := newFixture()
	o, replayed, err := f.svc.PlaceOrder(context.Background(), request("key-00000001", Line{ProductID: 1, Quantity: 2}, Line{ProductID: 2, Quantity: 1}))
	if err != nil {
		t.Fatal(err)
	}
	if replayed || o.Status != StatusPaid || o.TotalCents != 2*8999+1299 || o.CardLast4 != "4242" {
		t.Fatalf("order %+v", o)
	}
	if f.products.items[1].Stock != 8 || f.products.items[2].Stock != 0 {
		t.Fatal("stock not reserved")
	}
}

func TestPlaceOrderMergesRepeatedLines(t *testing.T) {
	f := newFixture()
	o, _, err := f.svc.PlaceOrder(context.Background(), request("key-00000001", Line{ProductID: 1, Quantity: 1}, Line{ProductID: 1, Quantity: 2}))
	if err != nil {
		t.Fatal(err)
	}
	if len(o.Items) != 1 || o.Items[0].Quantity != 3 {
		t.Fatalf("items %+v", o.Items)
	}
}

func TestPlaceOrderDeclinedReleasesStock(t *testing.T) {
	f := newFixture()
	f.gateway.decline = true

	o, _, err := f.svc.PlaceOrder(context.Background(), request("key-00000001", Line{ProductID: 1, Quantity: 3}))
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != StatusFailed || o.FailureReason == "" {
		t.Fatalf("order %+v", o)
	}
	if f.products.items[1].Stock != 10 {
		t.Fatalf("stock not released: %d", f.products.items[1].Stock)
	}
}

func TestPlaceOrderInsufficientStock(t *testing.T) {
	f := newFixture()
	_, _, err := f.svc.PlaceOrder(context.Background(), request("key-00000001", Line{ProductID: 2, Quantity: 5}))
	if !errors.Is(err, catalog.ErrInsufficientStock) {
		t.Fatalf("got %v", err)
	}
	if f.gateway.calls != 0 {
		t.Fatal("payment should not be attempted")
	}
}

func TestPlaceOrderUnknownProduct(t *testing.T) {
	f := newFixture()
	_, _, err := f.svc.PlaceOrder(context.Background(), request("key-00000001", Line{ProductID: 404, Quantity: 1}))
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("got %v", err)
	}
}

func TestPlaceOrderReplaysSameKey(t *testing.T) {
	f := newFixture()
	req := request("key-00000001", Line{ProductID: 1, Quantity: 1})

	first, _, err := f.svc.PlaceOrder(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	second, replayed, err := f.svc.PlaceOrder(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !replayed || second.ID != first.ID || f.gateway.calls != 1 || f.products.items[1].Stock != 9 {
		t.Fatalf("replay charged twice or moved stock: calls=%d stock=%d", f.gateway.calls, f.products.items[1].Stock)
	}
}

func TestPlaceOrderRejectsKeyReuseWithDifferentBody(t *testing.T) {
	f := newFixture()
	_, _, _ = f.svc.PlaceOrder(context.Background(), request("key-00000001", Line{ProductID: 1, Quantity: 1}))

	_, _, err := f.svc.PlaceOrder(context.Background(), request("key-00000001", Line{ProductID: 1, Quantity: 2}))
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("got %v", err)
	}
}

func TestPlaceOrderValidation(t *testing.T) {
	tests := map[string]PlaceOrder{
		"missing key":   {Lines: []Line{{ProductID: 1, Quantity: 1}}, CardNumber: "4242424242424242"},
		"no lines":      {IdempotencyKey: "key-00000001", CardNumber: "4242424242424242"},
		"zero quantity": {IdempotencyKey: "key-00000001", Lines: []Line{{ProductID: 1, Quantity: 0}}, CardNumber: "4242424242424242"},
		"bad card":      {IdempotencyKey: "key-00000001", Lines: []Line{{ProductID: 1, Quantity: 1}}, CardNumber: "4242"},
	}
	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, err := newFixture().svc.PlaceOrder(context.Background(), req)
			var verr *ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestDeclineReasonIsReadable(t *testing.T) {
	f := newFixture()
	f.gateway.decline = true
	o, _, _ := f.svc.PlaceOrder(context.Background(), request("key-00000009", Line{ProductID: 1, Quantity: 1}))
	if o.FailureReason != "insufficient funds" {
		t.Fatalf("got %q", o.FailureReason)
	}
}
