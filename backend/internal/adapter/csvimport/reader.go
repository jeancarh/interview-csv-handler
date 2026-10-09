package csvimport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
)

var (
	ErrInvalidFile = errors.New("invalid csv file")
	ErrTooManyRows = errors.New("csv file has too many rows")
)

var columns = []string{"name", "sku", "description", "category", "price", "stock", "weight_kg"}

func Parse(r io.Reader, maxRows int) ([]catalog.ImportRow, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.ReuseRecord = true

	header, err := cr.Read()
	if errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: the file is empty", ErrInvalidFile)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFile, err)
	}
	index, err := columnIndex(header)
	if err != nil {
		return nil, err
	}

	var rows []catalog.ImportRow
	for {
		record, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return rows, nil
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidFile, err)
		}
		line, _ := cr.FieldPos(0)
		if len(record) != len(columns) {
			return nil, fmt.Errorf("%w: line %d has %d columns, expected %d", ErrInvalidFile, line, len(record), len(columns))
		}
		if len(rows) == maxRows {
			return nil, fmt.Errorf("%w: the limit is %d", ErrTooManyRows, maxRows)
		}
		rows = append(rows, catalog.ImportRow{
			Line:        line,
			Name:        record[index["name"]],
			SKU:         record[index["sku"]],
			Description: record[index["description"]],
			Category:    record[index["category"]],
			Price:       record[index["price"]],
			Stock:       record[index["stock"]],
			WeightKg:    record[index["weight_kg"]],
		})
	}
}

func columnIndex(header []string) (map[string]int, error) {
	index := make(map[string]int, len(header))
	for i, h := range header {
		name := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, "\uFEFF")))
		if _, dup := index[name]; dup {
			return nil, fmt.Errorf("%w: column %q appears twice", ErrInvalidFile, name)
		}
		index[name] = i
	}
	var missing []string
	for _, c := range columns {
		if _, ok := index[c]; !ok {
			missing = append(missing, c)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: missing columns %s", ErrInvalidFile, strings.Join(missing, ", "))
	}
	if len(header) != len(columns) {
		return nil, fmt.Errorf("%w: expected exactly these columns: %s", ErrInvalidFile, strings.Join(columns, ", "))
	}
	return index, nil
}
