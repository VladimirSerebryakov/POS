package memory_test

import (
	"slices"
	"testing"

	"pos/internal/adapters/memory"
	"pos/internal/domain"
)

func TestRubricRepositoryInitiallyHasNoRubric(t *testing.T) {
	rubric, err := memory.NewRubricRepository().Get()
	if err != nil || rubric != nil {
		t.Fatalf("Get() = %v, %v; want nil, nil", rubric, err)
	}
}

func TestRubricRepositoryCopiesInputAndOutput(t *testing.T) {
	repository := memory.NewRubricRepository()
	criterion := testCriterion(t, 1)
	original := testRubric(t, criterion)
	if err := repository.Save(original); err != nil {
		t.Fatal(err)
	}
	second := testCriterion(t, 2)
	if err := original.AddCriterion(&second); err != nil {
		t.Fatal(err)
	}
	stored, err := repository.Get()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(stored.Criteria(), []domain.Criterion{criterion}) {
		t.Error("changing the input modified the stored rubric")
	}
	if err := stored.AddCriterion(&second); err != nil {
		t.Fatal(err)
	}
	stored, err = repository.Get()
	if err != nil || !slices.Equal(stored.Criteria(), []domain.Criterion{criterion}) {
		t.Fatalf("changing the returned rubric modified storage: %v, %v", stored, err)
	}
	if err := repository.Save(testRubric(t, second)); err != nil {
		t.Fatal(err)
	}
	stored, err = repository.Get()
	if err != nil || !slices.Equal(stored.Criteria(), []domain.Criterion{second}) {
		t.Fatalf("Save did not replace the rubric: %v, %v", stored, err)
	}
}

func TestRubricRepositoryPreservesStateOnRejectedSave(t *testing.T) {
	repository := memory.NewRubricRepository()
	criterion := testCriterion(t, 1)
	if err := repository.Save(testRubric(t, criterion)); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []*domain.Rubric{nil, {}} {
		if err := repository.Save(invalid); err == nil {
			t.Error("Save must reject a nil or empty rubric")
		}
		stored, err := repository.Get()
		if err != nil || !slices.Equal(stored.Criteria(), []domain.Criterion{criterion}) {
			t.Fatalf("a failed save changed the rubric: %v, %v", stored, err)
		}
	}
}
