package memory_test

import (
	"slices"
	"testing"

	"pos/internal/adapters/memory"
	"pos/internal/domain"
)

func TestEvaluationRepositoryCopiesSnapshots(t *testing.T) {
	repository := memory.NewEvaluationRepository()
	first := testEvaluation(t, 1, 1, 0)
	second := testEvaluation(t, 2, 1, 3)
	third := testEvaluation(t, 1, 2, 5)
	want := []domain.Evaluation{first, second, third}
	input := slices.Clone(want)
	if err := repository.Save(input); err != nil {
		t.Fatal(err)
	}
	input[0] = domain.Evaluation{}
	stored, err := repository.List()
	if err != nil || !slices.Equal(stored, want) {
		t.Fatalf("changing the input changed storage: %v, %v", stored, err)
	}
	stored[1] = domain.Evaluation{}
	stored, err = repository.List()
	if err != nil || !slices.Equal(stored, want) {
		t.Fatalf("changing the output changed storage: %v, %v", stored, err)
	}
}

func TestEvaluationRepositoryRejectsInvalidSnapshotsAtomically(t *testing.T) {
	repository := memory.NewEvaluationRepository()
	original := testEvaluation(t, 1, 1, 0)
	if err := repository.Save([]domain.Evaluation{original}); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		proposed []domain.Evaluation
	}{
		{name: "duplicate zero score", proposed: []domain.Evaluation{original, original}},
		{name: "conflicting scores for one pair", proposed: []domain.Evaluation{original, testEvaluation(t, 1, 1, 5)}},
		{name: "zero-value evaluation after valid data", proposed: []domain.Evaluation{testEvaluation(t, 2, 2, 3), {}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := repository.Save(tt.proposed); err == nil {
				t.Error("Save must reject an invalid snapshot")
			}
			stored, err := repository.List()
			if err != nil || !slices.Equal(stored, []domain.Evaluation{original}) {
				t.Fatalf("a failed save changed existing evaluations: %v, %v", stored, err)
			}
		})
	}
}

func TestEvaluationRepositoryReplacesAndClearsSnapshot(t *testing.T) {
	repository := memory.NewEvaluationRepository()
	if err := repository.Save([]domain.Evaluation{testEvaluation(t, 1, 1, 3)}); err != nil {
		t.Fatal(err)
	}
	corrected := testEvaluation(t, 1, 1, 5)
	if err := repository.Save([]domain.Evaluation{corrected}); err != nil {
		t.Fatal(err)
	}
	stored, err := repository.List()
	if err != nil || !slices.Equal(stored, []domain.Evaluation{corrected}) {
		t.Fatalf("snapshot replacement did not persist correction: %v, %v", stored, err)
	}
	if err := repository.Save(nil); err != nil {
		t.Fatal(err)
	}
	stored, err = repository.List()
	if err != nil || len(stored) != 0 {
		t.Fatalf("saving an empty snapshot must clear evaluations: %v, %v", stored, err)
	}
}
