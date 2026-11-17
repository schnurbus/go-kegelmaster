package handlers

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
)

func TestValidateFeeQuantity(t *testing.T) {
	whole := penaltytype.PenaltyType{Name: "Pumpe"}
	decimal := penaltytype.PenaltyType{Name: "Rinne", AllowsDecimalQuantity: true}

	tests := []struct {
		name  string
		pt    penaltytype.PenaltyType
		count float64
		ok    bool
	}{
		{"whole number", whole, 3, true},
		{"fraction on whole type", whole, 1.5, false},
		{"tiny fraction on whole type", whole, 2.0000001, false},
		{"two decimals", decimal, 2.25, true},
		{"float noise is fine", decimal, 0.1 + 0.2, true},
		{"three decimals", decimal, 1.255, false},
		{"integer on decimal type", decimal, 4, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFeeQuantity(tt.pt, tt.count)
			if tt.ok && err != nil {
				t.Fatalf("expected ok, got %v", err)
			}
			if !tt.ok {
				var fe *fiber.Error
				if !errors.As(err, &fe) || fe.Code != fiber.StatusBadRequest {
					t.Fatalf("expected 400, got %v", err)
				}
			}
		})
	}
}
