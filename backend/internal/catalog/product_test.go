package catalog

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validDraft() Draft {
	w := int64(350)
	return Draft{
		SKU:         "RS-001",
		Name:        "Running Shoes",
		Description: "Lightweight running shoes",
		Category:    "Footwear",
		PriceCents:  8999,
		Stock:       150,
		WeightGrams: &w,
	}
}

func TestNewProductTrimsAndStamps(t *testing.T) {
	d := validDraft()
	d.Name = "  Running Shoes  "
	now := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)

	p, err := NewProduct(d, now)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Running Shoes" {
		t.Fatalf("name not trimmed: %q", p.Name)
	}
	if !p.CreatedAt.Equal(now) || !p.UpdatedAt.Equal(now) {
		t.Fatal("timestamps not set")
	}
}

func TestDraftValidation(t *testing.T) {
	tests := []struct {
		name  string
		edit  func(*Draft)
		field string
	}{
		{"empty name", func(d *Draft) { d.Name = "" }, "name"},
		{"whitespace name", func(d *Draft) { d.Name = "     " }, "name"},
		{"long name", func(d *Draft) { d.Name = strings.Repeat("a", maxNameLen+1) }, "name"},
		{"empty sku", func(d *Draft) { d.SKU = "" }, "sku"},
		{"sku with spaces", func(d *Draft) { d.SKU = "RS 001" }, "sku"},
		{"sku with quote", func(d *Draft) { d.SKU = "RS'001" }, "sku"},
		{"empty category", func(d *Draft) { d.Category = " " }, "category"},
		{"negative price", func(d *Draft) { d.PriceCents = -1 }, "priceCents"},
		{"negative stock", func(d *Draft) { d.Stock = -5 }, "stock"},
		{"negative weight", func(d *Draft) { w := int64(-1); d.WeightGrams = &w }, "weightGrams"},
		{"long description", func(d *Draft) { d.Description = strings.Repeat("a", maxDescriptionLen+1) }, "description"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := validDraft()
			tt.edit(&d)
			_, err := NewProduct(d, time.Now())
			var verr *ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("want ValidationError, got %v", err)
			}
			if _, ok := verr.Fields[tt.field]; !ok {
				t.Fatalf("want error on %s, got %v", tt.field, verr.Fields)
			}
		})
	}
}

func TestDraftAllowsZeroPriceAndMissingWeight(t *testing.T) {
	d := validDraft()
	d.PriceCents = 0
	d.WeightGrams = nil
	if _, err := NewProduct(d, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestDraftKeepsMarkupAsPlainText(t *testing.T) {
	d := validDraft()
	d.Name = "<script>alert('xss')</script>"
	p, err := NewProduct(d, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != d.Name {
		t.Fatalf("name was altered: %q", p.Name)
	}
}

func TestUpdateKeepsIdentity(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p, _ := NewProduct(validDraft(), created)
	p.ID = 7

	d := validDraft()
	d.Stock = 3
	later := created.Add(time.Hour)
	if err := p.Update(d, later); err != nil {
		t.Fatal(err)
	}
	if p.ID != 7 || !p.CreatedAt.Equal(created) || !p.UpdatedAt.Equal(later) || p.Stock != 3 {
		t.Fatalf("unexpected product after update: %+v", p)
	}
}

func TestMergeCatalogDataLeavesStockAlone(t *testing.T) {
	p, _ := NewProduct(validDraft(), time.Now())
	p.Stock = 145

	d := validDraft()
	d.PriceCents = 9499
	d.Stock = 120
	changed := p.MergeCatalogData(d, time.Now())

	if !changed {
		t.Fatal("expected a change")
	}
	if p.PriceCents != 9499 || p.Stock != 145 {
		t.Fatalf("got price %d stock %d", p.PriceCents, p.Stock)
	}
}

func TestMergeCatalogDataDetectsNoChange(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p, _ := NewProduct(validDraft(), created)

	d := validDraft()
	d.Stock = 1
	if p.MergeCatalogData(d, created.Add(time.Hour)) {
		t.Fatal("stock-only difference should not count as a change")
	}
	if !p.UpdatedAt.Equal(created) {
		t.Fatal("updated_at should not move when nothing changed")
	}
}

func TestRangeMessages(t *testing.T) {
	d := validDraft()
	d.Stock = -5
	d.PriceCents = maxPriceCents + 1
	_, err := NewProduct(d, time.Now())
	var verr *ValidationError
	errors.As(err, &verr)
	if verr.Fields["stock"] != "can't be negative" || verr.Fields["priceCents"] != "can't be more than 1,000,000.00" {
		t.Fatalf("got %v", verr.Fields)
	}
}
