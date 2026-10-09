package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/csvimport"
	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
)

const maxJSONBody = 1 << 20

func (a *api) searchProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := catalog.SearchQuery{Text: q.Get("q"), Category: q.Get("category")}
	var err error
	if query.MinPriceCents, err = optionalInt(q.Get("minPriceCents"), "minPriceCents"); err != nil {
		writeError(w, r, err)
		return
	}
	if query.MaxPriceCents, err = optionalInt(q.Get("maxPriceCents"), "maxPriceCents"); err != nil {
		writeError(w, r, err)
		return
	}
	query.InStockOnly = q.Get("inStock") == "true"
	query.Page, _ = strconv.Atoi(q.Get("page"))
	query.PageSize, _ = strconv.Atoi(q.Get("pageSize"))

	page, err := a.catalog.Search(r.Context(), query)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toProductPage(page))
}

func (a *api) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	p, err := a.catalog.Get(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toProduct(p))
}

func (a *api) createProduct(w http.ResponseWriter, r *http.Request) {
	var in productInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, r, err)
		return
	}
	p, err := a.catalog.Create(r.Context(), in.draft())
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/api/v1/products/%d", p.ID))
	writeJSON(w, http.StatusCreated, toProduct(p))
}

func (a *api) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var in productInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, r, err)
		return
	}
	p, err := a.catalog.Update(r.Context(), id, in.draft())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toProduct(p))
}

func (a *api) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if err := a.catalog.Delete(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := a.catalog.Categories(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string][]string{"items": categories})
}

func (a *api) importProducts(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxUploadBytes)
	file, _, err := r.FormFile("file")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, r, err)
			return
		}
		writeError(w, r, fmt.Errorf("%w: send the csv as multipart form field \"file\"", errBadRequest))
		return
	}
	defer file.Close()

	rows, err := csvimport.Parse(file, a.maxImportRows)
	if err != nil {
		writeError(w, r, err)
		return
	}
	report, err := a.importer.Import(r.Context(), rows)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toImportReport(report))
}

func (a *api) placeOrder(w http.ResponseWriter, r *http.Request) {
	var in orderInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, r, err)
		return
	}
	req := checkout.PlaceOrder{IdempotencyKey: r.Header.Get("Idempotency-Key"), CardNumber: in.CardNumber}
	for _, it := range in.Items {
		req.Lines = append(req.Lines, checkout.Line{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	order, replayed, err := a.checkout.PlaceOrder(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	status := http.StatusCreated
	if replayed {
		w.Header().Set("Idempotent-Replayed", "true")
		status = http.StatusOK
	}
	w.Header().Set("Location", fmt.Sprintf("/api/v1/orders/%d", order.ID))
	writeJSON(w, status, toOrder(order))
}

func (a *api) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	order, err := a.checkout.Get(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toOrder(order))
}

func (a *api) health(w http.ResponseWriter, r *http.Request) {
	if err := a.ping(r.Context()); err != nil {
		writeProblem(w, http.StatusServiceUnavailable, "database is not reachable", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return err
		}
		return fmt.Errorf("%w: invalid json body: %v", errBadRequest, err)
	}
	return nil
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("%w: id must be a positive number", errBadRequest)
	}
	return id, nil
}

func optionalInt(raw, name string) (*int64, error) {
	if raw == "" {
		return nil, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %s must be a whole number of cents", errBadRequest, name)
	}
	return &n, nil
}
