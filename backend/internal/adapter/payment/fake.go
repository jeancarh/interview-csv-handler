package payment

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
)

const declineSuffix = "0002"

type Fake struct{}

func NewFake() Fake {
	return Fake{}
}

func (Fake) Charge(_ context.Context, c checkout.Charge) error {
	if !luhnValid(c.CardNumber) {
		return fmt.Errorf("%w: card number is not valid", checkout.ErrPaymentDeclined)
	}
	if strings.HasSuffix(c.CardNumber, declineSuffix) {
		return fmt.Errorf("%w: card was declined by the issuer", checkout.ErrPaymentDeclined)
	}
	return nil
}

func luhnValid(number string) bool {
	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if d < 0 || d > 9 {
			return false
		}
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return len(number) > 0 && sum%10 == 0
}
