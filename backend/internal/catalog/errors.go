package catalog

import (
	"errors"
	"sort"
	"strings"
)

var (
	ErrNotFound          = errors.New("product not found")
	ErrDuplicateSKU      = errors.New("a product with this sku already exists")
	ErrInsufficientStock = errors.New("not enough stock")
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
	return "invalid product: " + strings.Join(parts, "; ")
}
