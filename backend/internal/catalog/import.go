package catalog

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type RowStatus string

const (
	StatusCreated   RowStatus = "created"
	StatusUpdated   RowStatus = "updated"
	StatusUnchanged RowStatus = "unchanged"
	StatusSkipped   RowStatus = "skipped"
	StatusRejected  RowStatus = "rejected"
)

var markup = regexp.MustCompile(`<\s*/?\s*[a-zA-Z!][^>]*>`)

var csvColumn = map[string]string{
	"sku":         "sku",
	"name":        "name",
	"description": "description",
	"category":    "category",
	"priceCents":  "price",
	"stock":       "stock",
	"weightGrams": "weight_kg",
}

type ImportRow struct {
	Line        int
	Name        string
	SKU         string
	Description string
	Category    string
	Price       string
	Stock       string
	WeightKg    string
}

type RowResult struct {
	Line     int
	SKU      string
	Name     string
	Status   RowStatus
	Messages []string
}

type ImportSummary struct {
	Total     int
	Created   int
	Updated   int
	Unchanged int
	Skipped   int
	Rejected  int
}

type ImportReport struct {
	Summary ImportSummary
	Rows    []RowResult
}

type Importer struct {
	repo Repository
	tx   Transactor
	now  func() time.Time
}

func NewImporter(repo Repository, tx Transactor, now func() time.Time) *Importer {
	return &Importer{repo: repo, tx: tx, now: now}
}

type parsedRow struct {
	result *RowResult
	draft  Draft
}

func (im *Importer) Import(ctx context.Context, rows []ImportRow) (ImportReport, error) {
	results := make([]RowResult, len(rows))
	var valid []parsedRow

	for i, r := range rows {
		results[i] = RowResult{Line: r.Line, SKU: strings.TrimSpace(r.SKU), Name: strings.TrimSpace(r.Name)}
		res := &results[i]

		if r.isBlank() {
			res.Status = StatusSkipped
			res.Messages = []string{"empty row"}
			continue
		}
		d, notes, problems := r.toDraft()
		if len(problems) > 0 {
			res.Status = StatusRejected
			res.Messages = problems
			continue
		}
		res.Messages = notes
		valid = append(valid, parsedRow{result: res, draft: d})
	}

	winners := resolveDuplicates(valid)

	err := im.tx.WithinTx(ctx, func(ctx context.Context) error {
		for _, w := range winners {
			status, err := im.save(ctx, w.draft)
			if err != nil {
				return fmt.Errorf("line %d: %w", w.result.Line, err)
			}
			w.result.Status = status
		}
		return nil
	})
	if err != nil {
		return ImportReport{}, err
	}

	return ImportReport{Summary: summarize(results), Rows: results}, nil
}

func (im *Importer) save(ctx context.Context, d Draft) (RowStatus, error) {
	existing, err := im.repo.FindBySKU(ctx, d.SKU)
	if errors.Is(err, ErrNotFound) {
		p, err := NewProduct(d, im.now())
		if err != nil {
			return "", err
		}
		return StatusCreated, im.repo.Create(ctx, &p)
	}
	if err != nil {
		return "", err
	}
	if !existing.MergeCatalogData(d, im.now()) {
		return StatusUnchanged, nil
	}
	return StatusUpdated, im.repo.Update(ctx, &existing)
}

func resolveDuplicates(rows []parsedRow) []parsedRow {
	last := map[string]parsedRow{}
	for _, r := range rows {
		last[r.draft.SKU] = r
	}

	winners := make([]parsedRow, 0, len(last))
	for _, r := range rows {
		w := last[r.draft.SKU]
		if w.result == r.result {
			winners = append(winners, r)
			continue
		}
		r.result.Status = StatusSkipped
		if sameDraft(r.draft, w.draft) {
			r.result.Messages = append(r.result.Messages, fmt.Sprintf("same as line %d, which is the one imported", w.result.Line))
		} else {
			r.result.Messages = append(r.result.Messages, fmt.Sprintf("sku repeated later in the file, line %d wins", w.result.Line))
		}
	}
	return winners
}

func sameDraft(a, b Draft) bool {
	return a.SKU == b.SKU && a.Name == b.Name && a.Description == b.Description &&
		a.Category == b.Category && a.PriceCents == b.PriceCents && a.Stock == b.Stock &&
		equalWeight(a.WeightGrams, b.WeightGrams)
}

func (r ImportRow) isBlank() bool {
	for _, v := range []string{r.Name, r.SKU, r.Description, r.Category, r.Price, r.Stock, r.WeightKg} {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func (r ImportRow) toDraft() (Draft, []string, []string) {
	var notes, problems []string
	d := Draft{SKU: r.SKU, Name: r.Name, Description: r.Description, Category: r.Category}

	price := strings.TrimSpace(r.Price)
	if p, ok := strings.CutPrefix(price, "$"); ok {
		price = strings.TrimSpace(p)
		notes = append(notes, "removed currency symbol from price")
	}
	if cents, err := parseFixed(price, 2); err != nil {
		problems = append(problems, fmt.Sprintf("price %q is not a valid amount", r.Price))
	} else {
		d.PriceCents = cents
	}

	if stock, err := strconv.ParseInt(strings.TrimSpace(r.Stock), 10, 64); err != nil {
		problems = append(problems, fmt.Sprintf("stock %q is not a whole number", r.Stock))
	} else {
		d.Stock = stock
	}

	if w := strings.TrimSpace(r.WeightKg); w == "" {
		notes = append(notes, "no weight given, saved without weight")
	} else if grams, err := parseFixed(w, 3); err != nil {
		problems = append(problems, fmt.Sprintf("weight_kg %q is not a valid number", r.WeightKg))
	} else {
		d.WeightGrams = &grams
	}

	for _, f := range []struct{ name, value string }{{"name", r.Name}, {"description", r.Description}} {
		if markup.MatchString(f.value) {
			notes = append(notes, f.name+" contains HTML markup, stored as plain text")
		}
	}

	d, err := d.normalized()
	var verr *ValidationError
	if errors.As(err, &verr) {
		problems = append(problems, fieldProblems(verr)...)
	}
	return d, notes, problems
}

func fieldProblems(verr *ValidationError) []string {
	out := make([]string, 0, len(verr.Fields))
	for field, msg := range verr.Fields {
		out = append(out, csvColumn[field]+" "+msg)
	}
	sort.Strings(out)
	return out
}

func summarize(rows []RowResult) ImportSummary {
	s := ImportSummary{Total: len(rows)}
	for _, r := range rows {
		switch r.Status {
		case StatusCreated:
			s.Created++
		case StatusUpdated:
			s.Updated++
		case StatusUnchanged:
			s.Unchanged++
		case StatusSkipped:
			s.Skipped++
		case StatusRejected:
			s.Rejected++
		}
	}
	return s
}
