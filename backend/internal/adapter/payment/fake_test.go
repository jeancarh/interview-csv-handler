package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
)

func TestFakeCharge(t *testing.T) {
	tests := map[string]bool{
		"4242424242424242": true,
		"5555555555554444": true,
		"4000000000000002": false,
		"4242424242424241": false,
	}
	for card, approved := range tests {
		err := NewFake().Charge(context.Background(), checkout.Charge{CardNumber: card, AmountCents: 100})
		if approved && err != nil {
			t.Fatalf("%s: unexpected error %v", card, err)
		}
		if !approved && !errors.Is(err, checkout.ErrPaymentDeclined) {
			t.Fatalf("%s: want decline, got %v", card, err)
		}
	}
}
