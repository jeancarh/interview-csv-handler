package catalog

import "testing"

func TestParseFixed(t *testing.T) {
	tests := []struct {
		in      string
		scale   int
		want    int64
		wantErr bool
	}{
		{"89.99", 2, 8999, false},
		{"18.75", 2, 1875, false},
		{"22", 2, 2200, false},
		{"22.5", 2, 2250, false},
		{"0.00", 2, 0, false},
		{"0.35", 3, 350, false},
		{"35.0", 3, 35000, false},
		{"0.03", 3, 30, false},
		{"1.999", 2, 0, true},
		{"free", 2, 0, true},
		{"", 2, 0, true},
		{"-1.00", 2, 0, true},
		{"1.2.3", 2, 0, true},
		{".5", 2, 0, true},
		{"5.", 2, 0, true},
		{"1e3", 2, 0, true},
		{"99999999999999999999", 2, 0, true},
	}
	for _, tt := range tests {
		got, err := parseFixed(tt.in, tt.scale)
		if (err != nil) != tt.wantErr {
			t.Fatalf("parseFixed(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
		}
		if got != tt.want {
			t.Fatalf("parseFixed(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
