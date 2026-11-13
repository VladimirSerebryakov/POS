package domain_test

import (
	"testing"

	"pos/internal/domain"
)

func TestNewCriterion(t *testing.T) {
	weight, err := domain.NewWeight(3)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := domain.NewCriterion(1, "  Ходил на пары \t", weight, 10)
	if err != nil {
		t.Fatal(err)
	}
	if criterion.ID() != 1 {
		t.Errorf("ID() = %d, want 1", criterion.ID())
	}
	if criterion.Name() != "Ходил на пары" {
		t.Errorf("Name() = %q, want %q", criterion.Name(), "Ходил на пары")
	}
	if criterion.Weight() != weight {
		t.Errorf("Weight() = %v, want %v", criterion.Weight(), weight)
	}
	if criterion.MaxValue() != 10 {
		t.Errorf("MaxValue() = %d, want 10", criterion.MaxValue())
	}
}

func TestNewCriterionAcceptsMinimumMaximum(t *testing.T) {
	weight, err := domain.NewWeight(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := domain.NewCriterion(1, "Задавал вопросы", weight, 1); err != nil {
		t.Fatalf("a maximum of 1 must be allowed: %v", err)
	}
}

func TestNewCriterionRejectsInvalidValues(t *testing.T) {
	weight, err := domain.NewWeight(3)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		id       int
		title    string
		weight   domain.Weight
		maxValue int
	}{
		{name: "zero ID", id: 0, title: "Ходил на пары", weight: weight, maxValue: 10},
		{name: "negative ID", id: -1, title: "Ходил на пары", weight: weight, maxValue: 10},
		{name: "empty name", id: 1, title: "", weight: weight, maxValue: 10},
		{name: "zero maximum", id: 1, title: "Ходил на пары", weight: weight, maxValue: 0},
		{name: "negative maximum", id: 1, title: "Ходил на пары", weight: weight, maxValue: -1},
		{name: "zero-value weight", id: 1, title: "Ходил на пары", weight: domain.Weight{}, maxValue: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criterion, err := domain.NewCriterion(tt.id, tt.title, tt.weight, tt.maxValue)
			if err == nil {
				t.Fatal("NewCriterion must reject invalid data")
			}
			if criterion != nil {
				t.Error("NewCriterion must not return a criterion on error")
			}
		})
	}
}
