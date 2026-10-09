package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/csvimport"
	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
	"github.com/jeancarh/interview-csv-handler/backend/migrations"
)

func clock() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := Migrate(context.Background(), db, migrations.Files); err != nil {
		t.Fatal(err)
	}
	return db
}

func importSample(t *testing.T, db *sql.DB) catalog.ImportReport {
	t.Helper()
	f, err := os.Open("../../../../testdata/products.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csvimport.Parse(f, 1000)
	if err != nil {
		t.Fatal(err)
	}
	report, err := catalog.NewImporter(NewProductRepository(db), NewTransactor(db), clock).Import(context.Background(), rows)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func TestMigrateIsRepeatable(t *testing.T) {
	db := openTestDB(t)
	if err := Migrate(context.Background(), db, migrations.Files); err != nil {
		t.Fatal(err)
	}
}

func TestImportSampleFile(t *testing.T) {
	db := openTestDB(t)

	first := importSample(t, db)
	want := catalog.ImportSummary{Total: 97, Created: 87, Skipped: 5, Rejected: 5}
	if first.Summary != want {
		t.Fatalf("first import: got %+v want %+v", first.Summary, want)
	}

	second := importSample(t, db)
	want = catalog.ImportSummary{Total: 97, Unchanged: 87, Skipped: 5, Rejected: 5}
	if second.Summary != want {
		t.Fatalf("second import: got %+v want %+v", second.Summary, want)
	}

	repo := NewProductRepository(db)
	shoes, err := repo.FindBySKU(context.Background(), "RS-001")
	if err != nil || shoes.PriceCents != 9499 {
		t.Fatalf("RS-001 should keep the last row: %+v %v", shoes, err)
	}
	xss, err := repo.FindBySKU(context.Background(), "XS-001")
	if err != nil || xss.Name != "<script>alert('xss')</script>" {
		t.Fatalf("XS-001 should be stored verbatim: %+v %v", xss, err)
	}
	var tables int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name = 'products'`).Scan(&tables); err != nil || tables != 1 {
		t.Fatal("products table should still exist")
	}
}

func TestSearch(t *testing.T) {
	db := openTestDB(t)
	importSample(t, db)
	repo := NewProductRepository(db)
	ctx := context.Background()

	tests := []struct {
		name  string
		query catalog.SearchQuery
		total int
	}{
		{"prefix text", catalog.SearchQuery{Text: "speak"}, 3},
		{"sku", catalog.SearchQuery{Text: "RS-050"}, 1},
		{"description", catalog.SearchQuery{Text: "waterproof"}, 5},
		{"category", catalog.SearchQuery{Category: "Kitchen"}, 5},
		{"text and category", catalog.SearchQuery{Text: "speaker", Category: "Electronics"}, 3},
		{"fts syntax is neutralised", catalog.SearchQuery{Text: `"); DROP TABLE products;--`}, 1},
		{"only punctuation", catalog.SearchQuery{Text: "-- ;"}, 87},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.query.Page, tt.query.PageSize = 1, 100
			page, err := repo.Search(ctx, tt.query)
			if err != nil {
				t.Fatal(err)
			}
			if page.Total != tt.total {
				names := []string{}
				for _, p := range page.Items {
					names = append(names, p.Name)
				}
				t.Fatalf("got %d %v", page.Total, names)
			}
		})
	}
}

func TestSearchPriceRangeAndPaging(t *testing.T) {
	db := openTestDB(t)
	importSample(t, db)
	repo := NewProductRepository(db)

	min, max := int64(10000), int64(20000)
	page, err := repo.Search(context.Background(), catalog.SearchQuery{MinPriceCents: &min, MaxPriceCents: &max, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 7 || len(page.Items) != 2 {
		t.Fatalf("total %d items %d", page.Total, len(page.Items))
	}
}

func TestSoftDeleteFreesSKU(t *testing.T) {
	db := openTestDB(t)
	svc := catalog.NewService(NewProductRepository(db), clock)
	ctx := context.Background()
	d := catalog.Draft{SKU: "AB-1", Name: "Thing", Category: "Misc", PriceCents: 100, Stock: 1}

	p, err := svc.Create(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(ctx, d); !errors.Is(err, catalog.ErrDuplicateSKU) {
		t.Fatalf("got %v", err)
	}
	if err := svc.Delete(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctx, p.ID); !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
	if _, err := svc.Create(ctx, d); err != nil {
		t.Fatal(err)
	}
}

func TestReserveStockNeverOversells(t *testing.T) {
	db := openTestDB(t)
	repo := NewProductRepository(db)
	ctx := context.Background()
	p, _ := catalog.NewProduct(catalog.Draft{SKU: "AB-1", Name: "Thing", Category: "Misc", PriceCents: 100, Stock: 5}, clock())
	if err := repo.Create(ctx, &p); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := repo.ReserveStock(ctx, p.ID, 1); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	got, _ := repo.Get(ctx, p.ID)
	if ok != 5 || got.Stock != 0 {
		t.Fatalf("reserved %d, stock left %d", ok, got.Stock)
	}
}

type approveAll struct{}

func (approveAll) Charge(context.Context, checkout.Charge) error { return nil }

func TestCheckoutEndToEnd(t *testing.T) {
	db := openTestDB(t)
	importSample(t, db)
	products := NewProductRepository(db)
	svc := checkout.NewService(products, NewOrderRepository(db), approveAll{}, NewTransactor(db), clock)
	ctx := context.Background()

	shoes, _ := products.FindBySKU(ctx, "RS-001")
	req := checkout.PlaceOrder{
		IdempotencyKey: "order-key-1",
		Lines:          []checkout.Line{{ProductID: shoes.ID, Quantity: 2}},
		CardNumber:     "4242424242424242",
	}

	order, _, err := svc.PlaceOrder(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := svc.Get(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Status != checkout.StatusPaid || loaded.TotalCents != 18998 || len(loaded.Items) != 1 || loaded.Items[0].Name != "Running Shoes" {
		t.Fatalf("order %+v", loaded)
	}

	_, replayed, err := svc.PlaceOrder(ctx, req)
	if err != nil || !replayed {
		t.Fatalf("replay: %v %v", replayed, err)
	}
	after, _ := products.Get(ctx, shoes.ID)
	if after.Stock != shoes.Stock-2 {
		t.Fatalf("stock %d, want %d", after.Stock, shoes.Stock-2)
	}
}

func TestReserveRollsBackWholeOrder(t *testing.T) {
	db := openTestDB(t)
	importSample(t, db)
	products := NewProductRepository(db)
	svc := checkout.NewService(products, NewOrderRepository(db), approveAll{}, NewTransactor(db), clock)
	ctx := context.Background()

	shoes, _ := products.FindBySKU(ctx, "RS-001")
	clock, _ := products.FindBySKU(ctx, "VC-001")

	_, _, err := svc.PlaceOrder(ctx, checkout.PlaceOrder{
		IdempotencyKey: "order-key-2",
		Lines:          []checkout.Line{{ProductID: shoes.ID, Quantity: 1}, {ProductID: clock.ID, Quantity: 1}},
		CardNumber:     "4242424242424242",
	})
	if !errors.Is(err, catalog.ErrInsufficientStock) {
		t.Fatalf("got %v", err)
	}
	after, _ := products.Get(ctx, shoes.ID)
	if after.Stock != shoes.Stock {
		t.Fatalf("stock moved: %d -> %d", shoes.Stock, after.Stock)
	}
}
