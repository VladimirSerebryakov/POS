package domain_test

import (
	"strconv"
	"testing"

	"pos/internal/domain"
)

func newEvaluationTestCriterion(t *testing.T, id, maxValue int) domain.Criterion {
	t.Helper()
	weight, err := domain.NewWeight(3)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := domain.NewCriterion(id, "Задавал вопросы", weight, maxValue)
	if err != nil {
		t.Fatal(err)
	}
	return *criterion
}

func TestNewEvaluationAcceptsScoresWithinCriterionRange(t *testing.T) {
	criterion := newEvaluationTestCriterion(t, 1, 5)
	for _, score := range []int{0, 3, 5} {
		t.Run(strconv.Itoa(score), func(t *testing.T) {
			evaluation, err := domain.NewEvaluation(7, criterion, score)
			if err != nil {
				t.Fatal(err)
			}
			if evaluation.StudentID() != 7 || evaluation.CriterionID() != 1 || evaluation.Score() != score {
				t.Errorf("got evaluation (%d, %d, %d), want (7, 1, %d)",
					evaluation.StudentID(), evaluation.CriterionID(), evaluation.Score(), score)
			}
		})
	}
}

func TestNewEvaluationUsesSelectedCriterionMaximum(t *testing.T) {
	if _, err := domain.NewEvaluation(1, newEvaluationTestCriterion(t, 2, 10), 10); err != nil {
		t.Fatalf("criterion 2 allows a score of 10: %v", err)
	}
	if _, err := domain.NewEvaluation(1, newEvaluationTestCriterion(t, 1, 5), 10); err == nil {
		t.Error("criterion 1 has a maximum of 5 and must reject a score of 10")
	}
}

func TestNewEvaluationRejectsInvalidValues(t *testing.T) {
	criterion := newEvaluationTestCriterion(t, 1, 5)
	tests := []struct {
		name      string
		studentID int
		criterion domain.Criterion
		score     int
	}{
		{name: "zero student ID", studentID: 0, criterion: criterion, score: 3},
		{name: "negative student ID", studentID: -1, criterion: criterion, score: 3},
		{name: "zero-value criterion", studentID: 1, criterion: domain.Criterion{}, score: 3},
		{name: "negative score", studentID: 1, criterion: criterion, score: -1},
		{name: "score above maximum", studentID: 1, criterion: criterion, score: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluation, err := domain.NewEvaluation(tt.studentID, tt.criterion, tt.score)
			if err == nil {
				t.Fatal("NewEvaluation must reject invalid data")
			}
			if evaluation != nil {
				t.Error("NewEvaluation must not return an evaluation on error")
			}
		})
	}
}

func TestEvaluationValidateRejectsZeroValue(t *testing.T) {
	var evaluation domain.Evaluation
	if err := evaluation.Validate(newEvaluationTestCriterion(t, 1, 5)); err == nil {
		t.Error("Validate must reject a zero-value evaluation")
	}
}

func TestEvaluationValidateUsesSuppliedCriterion(t *testing.T) {
	evaluation, err := domain.NewEvaluation(1, newEvaluationTestCriterion(t, 1, 10), 8)
	if err != nil {
		t.Fatal(err)
	}
	if err := evaluation.Validate(newEvaluationTestCriterion(t, 1, 5)); err == nil {
		t.Error("Validate must reject a score above the supplied criterion's maximum")
	}
	if err := evaluation.Validate(newEvaluationTestCriterion(t, 2, 10)); err == nil {
		t.Error("Validate must reject a criterion with a different ID")
	}
}

func TestEvaluationCorrectScorePreservesIdentityAndOriginal(t *testing.T) {
	criterion := newEvaluationTestCriterion(t, 1, 5)
	original, err := domain.NewEvaluation(7, criterion, 2)
	if err != nil {
		t.Fatal(err)
	}
	corrected, err := original.CorrectScore(criterion, 5)
	if err != nil {
		t.Fatal(err)
	}
	if corrected.StudentID() != 7 || corrected.CriterionID() != 1 || corrected.Score() != 5 {
		t.Errorf("unexpected corrected evaluation: %v", corrected)
	}
	if original.Score() != 2 {
		t.Error("CorrectScore modified the original evaluation")
	}
	if _, err := original.CorrectScore(newEvaluationTestCriterion(t, 2, 5), 3); err == nil {
		t.Error("CorrectScore must not change the criterion identity")
	}
}
