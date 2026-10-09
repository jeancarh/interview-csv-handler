package csvimport

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParseSampleFile(t *testing.T) {
	f, err := os.Open("../../../../testdata/products.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	rows, err := Parse(f, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 97 {
		t.Fatalf("got %d rows", len(rows))
	}

	byLine := map[int]string{}
	for _, r := range rows {
		byLine[r.Line] = r.Name
	}
	if byLine[2] != "Running Shoes" {
		t.Fatalf("line 2: %q", byLine[2])
	}
	if byLine[53] != "Comma, In Product Name" {
		t.Fatalf("line 53: %q", byLine[53])
	}
	if byLine[59] != `Quote "Inside" Name` {
		t.Fatalf("line 59: %q", byLine[59])
	}
}

func TestParseAcceptsReorderedHeaderAndBOM(t *testing.T) {
	in := "\uFEFFSKU, Name ,description,category,price,stock,weight_kg\nAB-1,Thing,desc,Misc,1.00,2,0.1\n"
	rows, err := Parse(strings.NewReader(in), 10)
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].SKU != "AB-1" || rows[0].Name != "Thing" || rows[0].Line != 2 {
		t.Fatalf("got %+v", rows[0])
	}
}

func TestParseRejectsBadFiles(t *testing.T) {
	tests := map[string]string{
		"empty":            "",
		"missing column":   "name,sku,description,category,price,stock\n",
		"duplicate column": "name,name,sku,description,category,price,stock,weight_kg\n",
		"wrong arity":      "name,sku,description,category,price,stock,weight_kg\na,b,c\n",
		"broken quotes":    "name,sku,description,category,price,stock,weight_kg\n\"a,b,c,d,e,f,g\n",
	}
	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(in), 10)
			if !errors.Is(err, ErrInvalidFile) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestParseEnforcesRowLimit(t *testing.T) {
	in := "name,sku,description,category,price,stock,weight_kg\n" + strings.Repeat("a,b,c,d,1,1,1\n", 3)
	if _, err := Parse(strings.NewReader(in), 2); !errors.Is(err, ErrTooManyRows) {
		t.Fatalf("got %v", err)
	}
}
