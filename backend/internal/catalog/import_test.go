package catalog

import (
	"context"
	"strings"
	"testing"
)

func row(line int, name, sku, desc, cat, price, stock, weight string) ImportRow {
	return ImportRow{Line: line, Name: name, SKU: sku, Description: desc, Category: cat, Price: price, Stock: stock, WeightKg: weight}
}

func runImport(t *testing.T, repo *memRepo, rows ...ImportRow) ImportReport {
	t.Helper()
	report, err := NewImporter(repo, noTx{}, fixedClock).Import(context.Background(), rows)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func resultFor(t *testing.T, r ImportReport, line int) RowResult {
	t.Helper()
	for _, res := range r.Rows {
		if res.Line == line {
			return res
		}
	}
	t.Fatalf("no result for line %d", line)
	return RowResult{}
}

func hasMessage(res RowResult, part string) bool {
	for _, m := range res.Messages {
		if strings.Contains(m, part) {
			return true
		}
	}
	return false
}

func TestImportRejectsBadRows(t *testing.T) {
	report := runImport(t, newMemRepo(),
		row(2, "Yoga Mat", "YM-015", "mat", "Sports", "free", "200", "1.2"),
		row(3, "Desk Lamp", "DL-007", "lamp", "Home & Office", "45.50", "-5", "2.1"),
		row(4, "", "HD-099", "headphones", "Electronics", "149.99", "30", "0.25"),
		row(5, "     ", "WS-001", "whitespace", "Misc", "5.00", "10", "0.1"),
		row(6, "Gift Card", "GC-025", "card", "", "25.00", "99999", "0"),
		row(7, "Thing", "TH-1", "x", "Misc", "1.00", "ten", "0.1"),
	)

	cases := map[int]string{2: "price", 3: "stock", 4: "name", 5: "name", 6: "category", 7: "stock"}
	for line, part := range cases {
		res := resultFor(t, report, line)
		if res.Status != StatusRejected || !hasMessage(res, part) {
			t.Fatalf("line %d: got %s %v", line, res.Status, res.Messages)
		}
	}
	if report.Summary.Rejected != 6 || report.Summary.Created != 0 {
		t.Fatalf("summary %+v", report.Summary)
	}
}

func TestImportNormalizesAndNotes(t *testing.T) {
	repo := newMemRepo()
	report := runImport(t, repo,
		row(2, "Wireless Mouse", "WM-042", "mouse", "Electronics", "$29.99", "75", "0.12"),
		row(3, "Gaming Keyboard", "GK-088", "keyboard", "Electronics", "129.99", "45", ""),
		row(4, "Mystery Box", "MB-001", "box", "Gifts", "0.00", "50", "1.0"),
	)

	if report.Summary.Created != 3 {
		t.Fatalf("summary %+v", report.Summary)
	}
	if !hasMessage(resultFor(t, report, 2), "currency symbol") {
		t.Fatal("expected a note about the currency symbol")
	}
	if !hasMessage(resultFor(t, report, 3), "weight") {
		t.Fatal("expected a note about the missing weight")
	}

	mouse, _ := repo.FindBySKU(context.Background(), "WM-042")
	if mouse.PriceCents != 2999 || *mouse.WeightGrams != 120 {
		t.Fatalf("mouse %+v", mouse)
	}
	keyboard, _ := repo.FindBySKU(context.Background(), "GK-088")
	if keyboard.WeightGrams != nil {
		t.Fatal("keyboard should have no weight")
	}
}

func TestImportSkipsEmptyRows(t *testing.T) {
	report := runImport(t, newMemRepo(), row(2, "", "", "", "", "", "", ""))
	if res := resultFor(t, report, 2); res.Status != StatusSkipped {
		t.Fatalf("got %s", res.Status)
	}
}

func TestImportLastDuplicateWins(t *testing.T) {
	repo := newMemRepo()
	report := runImport(t, repo,
		row(2, "Running Shoes", "RS-001", "old", "Footwear", "89.99", "150", "0.35"),
		row(3, "Bluetooth Speaker", "BS-021", "portable", "Electronics", "59.99", "110", "0.8"),
		row(4, "Running Shoes", "RS-001", "updated", "Footwear", "94.99", "120", "0.35"),
		row(5, "Bluetooth Speaker", "BS-021", "different", "Electronics", "49.99", "200", "0.75"),
		row(6, "Bluetooth Speaker", "BS-021", "portable", "Electronics", "59.99", "110", "0.8"),
	)

	if res := resultFor(t, report, 2); res.Status != StatusSkipped || !hasMessage(res, "line 4") {
		t.Fatalf("line 2: %+v", res)
	}
	if res := resultFor(t, report, 3); res.Status != StatusSkipped || !hasMessage(res, "same as line 6") {
		t.Fatalf("line 3: %+v", res)
	}
	if res := resultFor(t, report, 5); res.Status != StatusSkipped || !hasMessage(res, "line 6") {
		t.Fatalf("line 5: %+v", res)
	}

	shoes, _ := repo.FindBySKU(context.Background(), "RS-001")
	if shoes.PriceCents != 9499 || shoes.Description != "updated" {
		t.Fatalf("shoes %+v", shoes)
	}
	if report.Summary.Created != 2 || report.Summary.Skipped != 3 {
		t.Fatalf("summary %+v", report.Summary)
	}
}

func TestReimportIsIdempotentAndKeepsStock(t *testing.T) {
	repo := newMemRepo()
	shoes := row(2, "Running Shoes", "RS-001", "shoes", "Footwear", "89.99", "150", "0.35")
	runImport(t, repo, shoes)

	p, _ := repo.FindBySKU(context.Background(), "RS-001")
	p.Stock = 145
	_ = repo.Update(context.Background(), &p)

	report := runImport(t, repo, shoes)
	if report.Summary.Unchanged != 1 {
		t.Fatalf("summary %+v", report.Summary)
	}

	cheaper := shoes
	cheaper.Price = "79.99"
	report = runImport(t, repo, cheaper)
	if report.Summary.Updated != 1 {
		t.Fatalf("summary %+v", report.Summary)
	}

	p, _ = repo.FindBySKU(context.Background(), "RS-001")
	if p.Stock != 145 || p.PriceCents != 7999 {
		t.Fatalf("got stock %d price %d", p.Stock, p.PriceCents)
	}
}

func TestImportKeepsHostileTextAsData(t *testing.T) {
	repo := newMemRepo()
	report := runImport(t, repo,
		row(2, "<script>alert('xss')</script>", "XS-001", "x", "Electronics", "19.99", "100", "0.1"),
		row(3, "Robert'); DROP TABLE products;--", "SQL-001", "x", "Games", "9.99", "50", "0.5"),
	)
	if report.Summary.Created != 2 {
		t.Fatalf("summary %+v", report.Summary)
	}
	if !hasMessage(resultFor(t, report, 2), "HTML markup") {
		t.Fatal("expected a note about markup in the name")
	}
	if hasMessage(resultFor(t, report, 3), "HTML markup") {
		t.Fatal("sql-looking text is not markup")
	}
}
