package httpapi

import (
	"time"

	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
)

type productInput struct {
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	PriceCents  int64  `json:"priceCents"`
	Stock       int64  `json:"stock"`
	WeightGrams *int64 `json:"weightGrams"`
}

func (in productInput) draft() catalog.Draft {
	return catalog.Draft{
		SKU:         in.SKU,
		Name:        in.Name,
		Description: in.Description,
		Category:    in.Category,
		PriceCents:  in.PriceCents,
		Stock:       in.Stock,
		WeightGrams: in.WeightGrams,
	}
}

type productResponse struct {
	ID          int64     `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	PriceCents  int64     `json:"priceCents"`
	Stock       int64     `json:"stock"`
	WeightGrams *int64    `json:"weightGrams"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toProduct(p catalog.Product) productResponse {
	return productResponse{
		ID:          p.ID,
		SKU:         p.SKU,
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		PriceCents:  p.PriceCents,
		Stock:       p.Stock,
		WeightGrams: p.WeightGrams,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

type productPage struct {
	Items    []productResponse `json:"items"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

func toProductPage(p catalog.Page) productPage {
	items := make([]productResponse, 0, len(p.Items))
	for _, it := range p.Items {
		items = append(items, toProduct(it))
	}
	return productPage{Items: items, Total: p.Total, Page: p.Page, PageSize: p.PageSize}
}

type importSummary struct {
	Total     int `json:"total"`
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Unchanged int `json:"unchanged"`
	Skipped   int `json:"skipped"`
	Rejected  int `json:"rejected"`
}

type importRow struct {
	Line     int      `json:"line"`
	SKU      string   `json:"sku"`
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	Messages []string `json:"messages"`
}

type importReport struct {
	Summary importSummary `json:"summary"`
	Rows    []importRow   `json:"rows"`
}

func toImportReport(r catalog.ImportReport) importReport {
	rows := make([]importRow, 0, len(r.Rows))
	for _, row := range r.Rows {
		msgs := row.Messages
		if msgs == nil {
			msgs = []string{}
		}
		rows = append(rows, importRow{Line: row.Line, SKU: row.SKU, Name: row.Name, Status: string(row.Status), Messages: msgs})
	}
	s := r.Summary
	return importReport{
		Summary: importSummary{Total: s.Total, Created: s.Created, Updated: s.Updated, Unchanged: s.Unchanged, Skipped: s.Skipped, Rejected: s.Rejected},
		Rows:    rows,
	}
}

type orderInput struct {
	Items []struct {
		ProductID int64 `json:"productId"`
		Quantity  int64 `json:"quantity"`
	} `json:"items"`
	CardNumber string `json:"cardNumber"`
}

type orderItemResponse struct {
	ProductID      int64  `json:"productId"`
	SKU            string `json:"sku"`
	Name           string `json:"name"`
	UnitPriceCents int64  `json:"unitPriceCents"`
	Quantity       int64  `json:"quantity"`
	SubtotalCents  int64  `json:"subtotalCents"`
}

type orderResponse struct {
	ID            int64               `json:"id"`
	Status        string              `json:"status"`
	Items         []orderItemResponse `json:"items"`
	TotalCents    int64               `json:"totalCents"`
	CardLast4     string              `json:"cardLast4"`
	FailureReason string              `json:"failureReason,omitempty"`
	CreatedAt     time.Time           `json:"createdAt"`
}

func toOrder(o checkout.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, orderItemResponse{
			ProductID:      it.ProductID,
			SKU:            it.SKU,
			Name:           it.Name,
			UnitPriceCents: it.UnitPriceCents,
			Quantity:       it.Quantity,
			SubtotalCents:  it.SubtotalCents(),
		})
	}
	return orderResponse{
		ID:            o.ID,
		Status:        string(o.Status),
		Items:         items,
		TotalCents:    o.TotalCents,
		CardLast4:     o.CardLast4,
		FailureReason: o.FailureReason,
		CreatedAt:     o.CreatedAt,
	}
}
