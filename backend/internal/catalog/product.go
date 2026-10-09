package catalog

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxNameLen        = 200
	maxSKULen         = 64
	maxCategoryLen    = 80
	maxDescriptionLen = 2000
	maxPriceCents     = 100_000_000
	maxStock          = 1_000_000
	maxWeightGrams    = 10_000_000
)

var skuPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

type Product struct {
	ID          int64
	SKU         string
	Name        string
	Description string
	Category    string
	PriceCents  int64
	Stock       int64
	WeightGrams *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Draft struct {
	SKU         string
	Name        string
	Description string
	Category    string
	PriceCents  int64
	Stock       int64
	WeightGrams *int64
}

func NewProduct(d Draft, now time.Time) (Product, error) {
	d, err := d.normalized()
	if err != nil {
		return Product{}, err
	}
	p := Product{CreatedAt: now}
	p.apply(d, now)
	p.Stock = d.Stock
	return p, nil
}

func (p *Product) Update(d Draft, now time.Time) error {
	d, err := d.normalized()
	if err != nil {
		return err
	}
	p.apply(d, now)
	p.Stock = d.Stock
	return nil
}

func (p *Product) MergeCatalogData(d Draft, now time.Time) bool {
	if p.SKU == d.SKU &&
		p.Name == d.Name &&
		p.Description == d.Description &&
		p.Category == d.Category &&
		p.PriceCents == d.PriceCents &&
		equalWeight(p.WeightGrams, d.WeightGrams) {
		return false
	}
	p.apply(d, now)
	return true
}

func (p *Product) apply(d Draft, now time.Time) {
	p.SKU = d.SKU
	p.Name = d.Name
	p.Description = d.Description
	p.Category = d.Category
	p.PriceCents = d.PriceCents
	p.WeightGrams = d.WeightGrams
	p.UpdatedAt = now
}

func (d Draft) normalized() (Draft, error) {
	d.SKU = strings.TrimSpace(d.SKU)
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
	d.Category = strings.TrimSpace(d.Category)

	fields := map[string]string{}
	checkText(fields, "name", d.Name, maxNameLen, true)
	checkText(fields, "category", d.Category, maxCategoryLen, true)
	checkText(fields, "description", d.Description, maxDescriptionLen, false)

	switch {
	case d.SKU == "":
		fields["sku"] = "is required"
	case utf8.RuneCountInString(d.SKU) > maxSKULen:
		fields["sku"] = "is too long"
	case !skuPattern.MatchString(d.SKU):
		fields["sku"] = "may only contain letters, digits, dashes and underscores"
	}

	checkRange(fields, "priceCents", d.PriceCents, maxPriceCents, "1,000,000.00")
	checkRange(fields, "stock", d.Stock, maxStock, "1,000,000")
	if d.WeightGrams != nil {
		checkRange(fields, "weightGrams", *d.WeightGrams, maxWeightGrams, "10,000 kg")
	}

	if len(fields) > 0 {
		return d, &ValidationError{Fields: fields}
	}
	return d, nil
}

func checkText(fields map[string]string, name, value string, max int, required bool) {
	switch {
	case required && value == "":
		fields[name] = "is required"
	case utf8.RuneCountInString(value) > max:
		fields[name] = "is too long"
	}
}

func checkRange(fields map[string]string, name string, value, max int64, maxLabel string) {
	switch {
	case value < 0:
		fields[name] = "can't be negative"
	case value > max:
		fields[name] = "can't be more than " + maxLabel
	}
}

func equalWeight(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
