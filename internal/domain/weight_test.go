package domain_test

import (
	"strconv"
	"testing"

	"pos/internal/domain"
)

func TestNewWeightAcceptsAllowedValues(t *testing.T) {
	for value := 1; value <= 5; value++ {
		t.Run(strconv.Itoa(value), func(t *testing.T) {
			weight, err := domain.NewWeight(value)
			if err != nil {
				t.Fatalf("NewWeight(%d): %v", value, err)
			}
			if weight.Value() != value {
				t.Errorf("Value() = %d, want %d", weight.Value(), value)
			}
		})
	}
}

func TestNewWeightRejectsOutOfRangeValues(t *testing.T) {
	for _, value := range []int{-1, 0, 6} {
		t.Run(strconv.Itoa(value), func(t *testing.T) {
			if _, err := domain.NewWeight(value); err == nil {
				t.Errorf("NewWeight(%d) must reject an out-of-range weight", value)
			}
		})
	}
}

func TestWeightsWithSameValueAreEqual(t *testing.T) {
	first, err := domain.NewWeight(3)
	if err != nil {
		t.Fatal(err)
	}
	second, err := domain.NewWeight(3)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Error("weights with the same value must be equal")
	}
}
