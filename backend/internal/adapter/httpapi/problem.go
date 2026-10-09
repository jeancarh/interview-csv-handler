package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jeancarh/interview-csv-handler/backend/internal/adapter/csvimport"
	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
)

type problem struct {
	Type   string            `json:"type"`
	Title  string            `json:"title"`
	Status int               `json:"status"`
	Detail string            `json:"detail,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

var errBadRequest = errors.New("bad request")

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, status int, detail string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
		Errors: fields,
	})
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		catalogInvalid  *catalog.ValidationError
		checkoutInvalid *checkout.ValidationError
		tooLarge        *http.MaxBytesError
	)
	switch {
	case errors.As(err, &catalogInvalid):
		writeProblem(w, http.StatusUnprocessableEntity, "some fields are not valid", catalogInvalid.Fields)
	case errors.As(err, &checkoutInvalid):
		writeProblem(w, http.StatusUnprocessableEntity, "some fields are not valid", checkoutInvalid.Fields)
	case errors.As(err, &tooLarge):
		writeProblem(w, http.StatusRequestEntityTooLarge, "the request body is too large", nil)
	case errors.Is(err, errBadRequest), errors.Is(err, csvimport.ErrInvalidFile), errors.Is(err, csvimport.ErrTooManyRows):
		writeProblem(w, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, catalog.ErrNotFound), errors.Is(err, checkout.ErrNotFound):
		writeProblem(w, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, catalog.ErrDuplicateSKU), errors.Is(err, catalog.ErrInsufficientStock):
		writeProblem(w, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, checkout.ErrIdempotencyConflict):
		writeProblem(w, http.StatusUnprocessableEntity, err.Error(), nil)
	default:
		slog.ErrorContext(r.Context(), "request failed", "error", err, "request_id", requestID(r.Context()))
		writeProblem(w, http.StatusInternalServerError, "something went wrong on our side", nil)
	}
}
