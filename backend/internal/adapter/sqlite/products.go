package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
)

const productColumns = `p.id, p.sku, p.name, p.description, p.category, p.price_cents, p.stock, p.weight_grams, p.created_at, p.updated_at`

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *catalog.Product) error {
	res, err := conn(ctx, r.db).ExecContext(ctx,
		`INSERT INTO products (sku, name, description, category, price_cents, stock, weight_grams, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.SKU, p.Name, p.Description, p.Category, p.PriceCents, p.Stock, nullableInt(p.WeightGrams),
		formatTime(p.CreatedAt), formatTime(p.UpdatedAt))
	if isUniqueViolation(err) {
		return catalog.ErrDuplicateSKU
	}
	if err != nil {
		return err
	}
	p.ID, err = res.LastInsertId()
	return err
}

func (r *ProductRepository) Update(ctx context.Context, p *catalog.Product) error {
	res, err := conn(ctx, r.db).ExecContext(ctx,
		`UPDATE products
		 SET sku = ?, name = ?, description = ?, category = ?, price_cents = ?, stock = ?, weight_grams = ?, updated_at = ?
		 WHERE id = ? AND deleted_at IS NULL`,
		p.SKU, p.Name, p.Description, p.Category, p.PriceCents, p.Stock, nullableInt(p.WeightGrams),
		formatTime(p.UpdatedAt), p.ID)
	if isUniqueViolation(err) {
		return catalog.ErrDuplicateSKU
	}
	return expectOneRow(res, err, catalog.ErrNotFound)
}

func (r *ProductRepository) Get(ctx context.Context, id int64) (catalog.Product, error) {
	row := conn(ctx, r.db).QueryRowContext(ctx,
		`SELECT `+productColumns+` FROM products p WHERE p.id = ? AND p.deleted_at IS NULL`, id)
	return scanProduct(row)
}

func (r *ProductRepository) FindBySKU(ctx context.Context, sku string) (catalog.Product, error) {
	row := conn(ctx, r.db).QueryRowContext(ctx,
		`SELECT `+productColumns+` FROM products p WHERE p.sku = ? AND p.deleted_at IS NULL`, sku)
	return scanProduct(row)
}

func (r *ProductRepository) Delete(ctx context.Context, id int64, at time.Time) error {
	res, err := conn(ctx, r.db).ExecContext(ctx,
		`UPDATE products SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		formatTime(at), formatTime(at), id)
	return expectOneRow(res, err, catalog.ErrNotFound)
}

func (r *ProductRepository) ReserveStock(ctx context.Context, id, qty int64) error {
	res, err := conn(ctx, r.db).ExecContext(ctx,
		`UPDATE products SET stock = stock - ? WHERE id = ? AND deleted_at IS NULL AND stock >= ?`,
		qty, id, qty)
	if err := expectOneRow(res, err, catalog.ErrInsufficientStock); err != nil {
		if errors.Is(err, catalog.ErrInsufficientStock) {
			if _, getErr := r.Get(ctx, id); getErr != nil {
				return getErr
			}
		}
		return err
	}
	return nil
}

func (r *ProductRepository) ReleaseStock(ctx context.Context, id, qty int64) error {
	res, err := conn(ctx, r.db).ExecContext(ctx,
		`UPDATE products SET stock = stock + ? WHERE id = ?`, qty, id)
	return expectOneRow(res, err, catalog.ErrNotFound)
}

func (r *ProductRepository) Categories(ctx context.Context) ([]string, error) {
	rows, err := conn(ctx, r.db).QueryContext(ctx,
		`SELECT DISTINCT category FROM products WHERE deleted_at IS NULL ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []string{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *ProductRepository) Search(ctx context.Context, q catalog.SearchQuery) (catalog.Page, error) {
	from := ` FROM products p`
	where := []string{"p.deleted_at IS NULL"}
	var args []any
	order := ` ORDER BY p.name, p.id`

	if match := ftsQuery(q.Text); match != "" {
		from += ` JOIN products_fts ON products_fts.rowid = p.id`
		where = append(where, "products_fts MATCH ?")
		args = append(args, match)
		order = ` ORDER BY products_fts.rank, p.id`
	}
	if q.Category != "" {
		where = append(where, "p.category = ?")
		args = append(args, q.Category)
	}
	if q.MinPriceCents != nil {
		where = append(where, "p.price_cents >= ?")
		args = append(args, *q.MinPriceCents)
	}
	if q.MaxPriceCents != nil {
		where = append(where, "p.price_cents <= ?")
		args = append(args, *q.MaxPriceCents)
	}
	if q.InStockOnly {
		where = append(where, "p.stock > 0")
	}
	filter := from + ` WHERE ` + strings.Join(where, " AND ")

	page := catalog.Page{Items: []catalog.Product{}, Page: q.Page, PageSize: q.PageSize}
	if err := conn(ctx, r.db).QueryRowContext(ctx, `SELECT count(*)`+filter, args...).Scan(&page.Total); err != nil {
		return catalog.Page{}, err
	}

	rows, err := conn(ctx, r.db).QueryContext(ctx,
		`SELECT `+productColumns+filter+order+` LIMIT ? OFFSET ?`,
		append(args, q.PageSize, (q.Page-1)*q.PageSize)...)
	if err != nil {
		return catalog.Page{}, err
	}
	defer rows.Close()

	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return catalog.Page{}, err
		}
		page.Items = append(page.Items, p)
	}
	return page, rows.Err()
}

func ftsQuery(text string) string {
	var terms []string
	for _, word := range strings.Fields(text) {
		word = strings.ReplaceAll(word, `"`, "")
		if strings.IndexFunc(word, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) < 0 {
			continue
		}
		terms = append(terms, `"`+word+`"*`)
	}
	return strings.Join(terms, " ")
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(s scanner) (catalog.Product, error) {
	var (
		p                catalog.Product
		weight           sql.NullInt64
		created, updated string
	)
	err := s.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.PriceCents, &p.Stock, &weight, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return catalog.Product{}, catalog.ErrNotFound
	}
	if err != nil {
		return catalog.Product{}, err
	}
	if weight.Valid {
		p.WeightGrams = &weight.Int64
	}
	if p.CreatedAt, err = parseTime(created); err != nil {
		return catalog.Product{}, err
	}
	if p.UpdatedAt, err = parseTime(updated); err != nil {
		return catalog.Product{}, err
	}
	return p, nil
}

func nullableInt(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func expectOneRow(res sql.Result, err error, none error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return none
	}
	return nil
}
