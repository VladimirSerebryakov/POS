package domain_test

import (
	"testing"

	"pos/internal/domain"
)

func newRubricTestCriterion(t *testing.T, id int, name string) *domain.Criterion {
	t.Helper()
	weight, err := domain.NewWeight(3)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := domain.NewCriterion(id, name, weight, 10)
	if err != nil {
		t.Fatal(err)
	}
	return criterion
}

func TestRubricAddCriterionPreservesOrder(t *testing.T) {
	var rubric domain.Rubric
	first := newRubricTestCriterion(t, 7, "Ходил на пары")
	second := newRubricTestCriterion(t, 2, "Ходил на пары")
	for _, criterion := range []*domain.Criterion{first, second} {
		if err := rubric.AddCriterion(criterion); err != nil {
			t.Fatalf("AddCriterion(%d): %v", criterion.ID(), err)
		}
	}
	got := rubric.Criteria()
	if len(got) != 2 {
		t.Fatalf("got %d criteria, want 2", len(got))
	}
	if got[0] != *first || got[1] != *second {
		t.Errorf("Criteria() = %v, want [%v %v]", got, *first, *second)
	}
}

func TestRubricAddCriterionRejectsInvalidInputWithoutChangingState(t *testing.T) {
	tests := []struct {
		name      string
		criterion *domain.Criterion
	}{
		{name: "nil criterion", criterion: nil},
		{name: "zero-value criterion", criterion: &domain.Criterion{}},
		{name: "duplicate ID", criterion: newRubricTestCriterion(t, 1, "Задавал вопросы")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rubric domain.Rubric
			original := newRubricTestCriterion(t, 1, "Ходил на пары")
			if err := rubric.AddCriterion(original); err != nil {
				t.Fatal(err)
			}
			if err := rubric.AddCriterion(tt.criterion); err == nil {
				t.Fatal("AddCriterion must reject invalid input")
			}
			got := rubric.Criteria()
			if len(got) != 1 || got[0] != *original {
				t.Errorf("failed addition changed the rubric: got %v, want [%v]", got, *original)
			}
		})
	}
}

func TestRubricFindCriterion(t *testing.T) {
	var rubric domain.Rubric
	criteria := []*domain.Criterion{
		newRubricTestCriterion(t, 1, "Ходил на пары"),
		newRubricTestCriterion(t, 2, "Задавал вопросы"),
	}
	for _, criterion := range criteria {
		if err := rubric.AddCriterion(criterion); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range criteria {
		got, err := rubric.FindCriterion(want.ID())
		if err != nil {
			t.Fatalf("FindCriterion(%d): %v", want.ID(), err)
		}
		if got != *want {
			t.Errorf("FindCriterion(%d) = %v, want %v", want.ID(), got, *want)
		}
	}
}

func TestRubricFindCriterionRejectsMissingID(t *testing.T) {
	for _, populated := range []bool{false, true} {
		name := "empty rubric"
		if populated {
			name = "populated rubric"
		}
		t.Run(name, func(t *testing.T) {
			var rubric domain.Rubric
			if populated {
				if err := rubric.AddCriterion(newRubricTestCriterion(t, 1, "Ходил на пары")); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := rubric.FindCriterion(99); err == nil {
				t.Error("FindCriterion must return an error for a missing ID")
			}
		})
	}
}

func TestRubricAddCriterionCopiesInput(t *testing.T) {
	var rubric domain.Rubric
	criterion := newRubricTestCriterion(t, 1, "Ходил на пары")
	want := *criterion
	if err := rubric.AddCriterion(criterion); err != nil {
		t.Fatal(err)
	}
	*criterion = *newRubricTestCriterion(t, 2, "Задавал вопросы")
	got, err := rubric.FindCriterion(want.ID())
	if err != nil {
		t.Fatalf("replacing the input changed the stored ID: %v", err)
	}
	if got != want {
		t.Errorf("replacing the input changed the stored criterion: got %v, want %v", got, want)
	}
}

func TestRubricCriteriaReturnsIndependentSlice(t *testing.T) {
	var rubric domain.Rubric
	criterion := newRubricTestCriterion(t, 1, "Ходил на пары")
	if err := rubric.AddCriterion(criterion); err != nil {
		t.Fatal(err)
	}
	criteria := rubric.Criteria()
	if len(criteria) != 1 {
		t.Fatalf("got %d criteria, want 1", len(criteria))
	}
	criteria[0] = domain.Criterion{}
	stored := rubric.Criteria()
	if len(stored) != 1 || stored[0] != *criterion {
		t.Errorf("changing the returned slice changed the rubric: got %v, want [%v]", stored, *criterion)
	}
}
