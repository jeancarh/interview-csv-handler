package catalog

import (
	"errors"
	"math"
	"strings"
)

var errInvalidDecimal = errors.New("invalid decimal")

func parseFixed(s string, scale int) (int64, error) {
	whole, frac, hasDot := strings.Cut(s, ".")
	if whole == "" || (hasDot && frac == "") || len(frac) > scale {
		return 0, errInvalidDecimal
	}
	if !isDigits(whole) || !isDigits(frac) {
		return 0, errInvalidDecimal
	}
	frac += strings.Repeat("0", scale-len(frac))

	var n int64
	for _, r := range whole + frac {
		d := int64(r - '0')
		if n > (math.MaxInt64-d)/10 {
			return 0, errInvalidDecimal
		}
		n = n*10 + d
	}
	return n, nil
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
