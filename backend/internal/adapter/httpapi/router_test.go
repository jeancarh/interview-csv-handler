package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/payment"
	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/sqlite"
	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
	"github.com/jeancarh/interview-csv-handler/backend/migrations"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := sqlite.Migrate(context.Background(), db, migrations.Files); err != nil {
		t.Fatal(err)
	}

	now := func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }
	products := sqlite.NewProductRepository(db)
	tx := sqlite.NewTransactor(db)

	srv := httptest.NewServer(NewRouter(Deps{
		Catalog:        catalog.NewService(products, now),
		Importer:       catalog.NewImporter(products, tx, now),
		Checkout:       checkout.NewService(products, sqlite.NewOrderRepository(db), payment.NewFake(), tx, now),
		Ping:           db.PingContext,
		Assets:         fstest.MapFS{"index.html": {Data: []byte("<html>shop</html>")}},
		Spec:           []byte("openapi: 3.1.0"),
		MaxUploadBytes: 1 << 20,
		MaxImportRows:  1000,
	}))
	t.Cleanup(srv.Close)
	return srv
}

func call(t *testing.T, srv *httptest.Server, method, path string, body any, headers map[string]string) (*http.Response, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, srv.URL+path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out := map[string]any{}
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res, out
}

func uploadCSV(t *testing.T, srv *httptest.Server, content []byte) (*http.Response, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "products.csv")
	_, _ = fw.Write(content)
	mw.Close()

	res, err := http.Post(srv.URL+"/api/v1/products/import", mw.FormDataContentType(), &buf)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out := map[string]any{}
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res, out
}

func shoes() map[string]any {
	return map[string]any{
		"sku": "RS-001", "name": "Running Shoes", "description": "Daily trainer",
		"category": "Footwear", "priceCents": 8999, "stock": 10, "weightGrams": 350,
	}
}

func TestProductCRUD(t *testing.T) {
	srv := newServer(t)

	res, created := call(t, srv, http.MethodPost, "/api/v1/products", shoes(), nil)
	if res.StatusCode != http.StatusCreated || res.Header.Get("Location") == "" {
		t.Fatalf("create: %d %v", res.StatusCode, created)
	}
	id := int(created["id"].(float64))
	path := "/api/v1/products/" + strconv.Itoa(id)

	update := shoes()
	update["name"] = "Trail Shoes"
	if res, body := call(t, srv, http.MethodPut, path, update, nil); res.StatusCode != http.StatusOK || body["name"] != "Trail Shoes" {
		t.Fatalf("update: %d %v", res.StatusCode, body)
	}
	if res, _ := call(t, srv, http.MethodDelete, path, nil, nil); res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: %d", res.StatusCode)
	}
	if res, body := call(t, srv, http.MethodGet, path, nil, nil); res.StatusCode != http.StatusNotFound || res.Header.Get("Content-Type") != "application/problem+json" {
		t.Fatalf("get after delete: %d %v", res.StatusCode, body)
	}
}

func TestProductValidationAndConflicts(t *testing.T) {
	srv := newServer(t)

	bad := shoes()
	bad["name"] = "  "
	bad["priceCents"] = -1
	res, body := call(t, srv, http.MethodPost, "/api/v1/products", bad, nil)
	errs, _ := body["errors"].(map[string]any)
	if res.StatusCode != http.StatusUnprocessableEntity || errs["name"] == nil || errs["priceCents"] == nil {
		t.Fatalf("validation: %d %v", res.StatusCode, body)
	}

	unknown := shoes()
	unknown["colour"] = "red"
	if res, _ := call(t, srv, http.MethodPost, "/api/v1/products", unknown, nil); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown field: %d", res.StatusCode)
	}

	call(t, srv, http.MethodPost, "/api/v1/products", shoes(), nil)
	if res, _ := call(t, srv, http.MethodPost, "/api/v1/products", shoes(), nil); res.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate sku: %d", res.StatusCode)
	}
	if res, _ := call(t, srv, http.MethodGet, "/api/v1/products/abc", nil, nil); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad id: %d", res.StatusCode)
	}
}

func TestImportAndSearch(t *testing.T) {
	srv := newServer(t)
	sample, err := os.ReadFile("../../../../testdata/products.csv")
	if err != nil {
		t.Fatal(err)
	}

	res, report := uploadCSV(t, srv, sample)
	summary, _ := report["summary"].(map[string]any)
	if res.StatusCode != http.StatusOK || summary["created"] != float64(87) || summary["rejected"] != float64(5) {
		t.Fatalf("import: %d %v", res.StatusCode, summary)
	}

	res, page := call(t, srv, http.MethodGet, "/api/v1/products?q=speaker&category=Electronics&pageSize=2", nil, nil)
	if res.StatusCode != http.StatusOK || page["total"] != float64(3) || len(page["items"].([]any)) != 2 {
		t.Fatalf("search: %d %v", res.StatusCode, page)
	}

	if res, _ := call(t, srv, http.MethodGet, "/api/v1/products?minPriceCents=abc", nil, nil); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad filter: %d", res.StatusCode)
	}

	res, cats := call(t, srv, http.MethodGet, "/api/v1/categories", nil, nil)
	if res.StatusCode != http.StatusOK || len(cats["items"].([]any)) == 0 {
		t.Fatalf("categories: %v", cats)
	}
}

func TestImportRejectsBrokenFile(t *testing.T) {
	srv := newServer(t)
	res, body := uploadCSV(t, srv, []byte("name,sku\nx,y\n"))
	if res.StatusCode != http.StatusBadRequest || !strings.Contains(body["detail"].(string), "missing columns") {
		t.Fatalf("got %d %v", res.StatusCode, body)
	}
}

func TestCheckoutFlow(t *testing.T) {
	srv := newServer(t)
	_, product := call(t, srv, http.MethodPost, "/api/v1/products", shoes(), nil)
	id := product["id"]

	order := map[string]any{"items": []map[string]any{{"productId": id, "quantity": 2}}, "cardNumber": "4242 4242 4242 4242"}
	key := map[string]string{"Idempotency-Key": "checkout-test-0001"}

	res, body := call(t, srv, http.MethodPost, "/api/v1/orders", order, key)
	if res.StatusCode != http.StatusCreated || body["status"] != "paid" || body["totalCents"] != float64(17998) {
		t.Fatalf("place: %d %v", res.StatusCode, body)
	}

	res, again := call(t, srv, http.MethodPost, "/api/v1/orders", order, key)
	if res.StatusCode != http.StatusOK || res.Header.Get("Idempotent-Replayed") != "true" || again["id"] != body["id"] {
		t.Fatalf("replay: %d %v", res.StatusCode, again)
	}

	order["cardNumber"] = "4000 0000 0000 0002"
	res, declined := call(t, srv, http.MethodPost, "/api/v1/orders", order, map[string]string{"Idempotency-Key": "checkout-test-0002"})
	if res.StatusCode != http.StatusCreated || declined["status"] != "failed" || declined["failureReason"] == "" {
		t.Fatalf("declined: %d %v", res.StatusCode, declined)
	}

	_, p := call(t, srv, http.MethodGet, "/api/v1/products/"+strconv.Itoa(int(id.(float64))), nil, nil)
	if p["stock"] != float64(8) {
		t.Fatalf("stock should be 8 after one paid and one declined order, got %v", p["stock"])
	}

	order["items"] = []map[string]any{{"productId": id, "quantity": 50}}
	order["cardNumber"] = "4242424242424242"
	if res, _ := call(t, srv, http.MethodPost, "/api/v1/orders", order, map[string]string{"Idempotency-Key": "checkout-test-0003"}); res.StatusCode != http.StatusConflict {
		t.Fatalf("oversell: %d", res.StatusCode)
	}

	if res, _ := call(t, srv, http.MethodPost, "/api/v1/orders", order, nil); res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("missing key: %d", res.StatusCode)
	}
}

func TestRoutingEdges(t *testing.T) {
	srv := newServer(t)

	res, err := http.Get(srv.URL + "/products/42")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), "shop") {
		t.Fatalf("spa fallback: %d %s", res.StatusCode, body)
	}

	if res, _ := call(t, srv, http.MethodGet, "/api/v1/nope", nil, nil); res.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown api route: %d", res.StatusCode)
	}
	if res, body := call(t, srv, http.MethodGet, "/healthz", nil, nil); res.StatusCode != http.StatusOK || body["status"] != "ok" {
		t.Fatalf("health: %d", res.StatusCode)
	}
	if res, _ := call(t, srv, http.MethodGet, "/healthz", nil, nil); res.Header.Get("X-Request-ID") == "" {
		t.Fatal("missing request id")
	}
}
