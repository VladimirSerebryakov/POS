package service_test

import (
	"slices"
	"testing"

	"pos/internal/domain"
	domainservice "pos/internal/domain/service"
)

func TestRubricServiceUpdate(t *testing.T) {
	first := newEvaluationTestCriterion(t, 1, 5)
	second := newEvaluationTestCriterion(t, 2, 10)
	original := testRubric(t, first, second)
	replacement := newWeightedTestCriterion(t, 1, 20, 5)
	current := []domain.Evaluation{testEvaluation(t, 1, second, 7)}
	updated, err := (domainservice.RubricService{}).Update(original, replacement, current)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(updated.Criteria(), []domain.Criterion{replacement, second}) {
		t.Errorf("unexpected updated rubric: %v", updated.Criteria())
	}
	if !slices.Equal(original.Criteria(), []domain.Criterion{first, second}) {
		t.Error("Update modified the input rubric")
	}
	extra := newEvaluationTestCriterion(t, 3, 10)
	if err := updated.AddCriterion(&extra); err != nil {
		t.Fatal(err)
	}
	if len(original.Criteria()) != 2 {
		t.Error("updated rubric shares mutable storage with the input")
	}
}

func TestRubricServiceRemove(t *testing.T) {
	first := newEvaluationTestCriterion(t, 1, 5)
	second := newEvaluationTestCriterion(t, 2, 10)
	third := newEvaluationTestCriterion(t, 3, 10)
	original := testRubric(t, first, second, third)
	current := []domain.Evaluation{testEvaluation(t, 1, first, 3)}
	result, err := (domainservice.RubricService{}).Remove(original, 2, current)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(result.Criteria(), []domain.Criterion{first, third}) {
		t.Errorf("unexpected remaining criteria: %v", result.Criteria())
	}
	if !slices.Equal(original.Criteria(), []domain.Criterion{first, second, third}) {
		t.Error("Remove modified the input rubric")
	}
}

func TestRubricServiceRejectsChangesToEvaluatedCriterion(t *testing.T) {
	criterion := newEvaluationTestCriterion(t, 1, 5)
	other := newEvaluationTestCriterion(t, 2, 10)
	rubric := testRubric(t, criterion, other)
	current := []domain.Evaluation{testEvaluation(t, 99, criterion, 0)}
	service := domainservice.RubricService{}
	if updated, err := service.Update(rubric, newWeightedTestCriterion(t, 1, 20, 5), current); err == nil || updated != nil {
		t.Error("Update must reject a criterion with an evaluation")
	}
	if updated, err := service.Remove(rubric, 1, current); err == nil || updated != nil {
		t.Error("Remove must reject a criterion with an evaluation")
	}
	if !slices.Equal(rubric.Criteria(), []domain.Criterion{criterion, other}) {
		t.Error("a rejected change modified the rubric")
	}
}

func TestRubricServiceRejectsInvalidRequests(t *testing.T) {
	criterion := newEvaluationTestCriterion(t, 1, 5)
	tests := []struct {
		name      string
		rubric    *domain.Rubric
		criterion domain.Criterion
	}{
		{name: "nil rubric", criterion: criterion},
		{name: "empty rubric", rubric: &domain.Rubric{}, criterion: criterion},
		{name: "missing criterion", rubric: testRubric(t, newEvaluationTestCriterion(t, 2, 10)), criterion: criterion},
		{name: "invalid criterion", rubric: testRubric(t, criterion), criterion: domain.Criterion{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := domainservice.RubricService{}
			if result, err := service.Update(tt.rubric, tt.criterion, nil); err == nil || result != nil {
				t.Error("Update must reject an invalid request")
			}
			if result, err := service.Remove(tt.rubric, tt.criterion.ID(), nil); err == nil || result != nil {
				t.Error("Remove must reject an invalid request")
			}
		})
	}
}

func TestRubricServiceCannotRemoveLastCriterion(t *testing.T) {
	criterion := newEvaluationTestCriterion(t, 1, 5)
	rubric := testRubric(t, criterion)
	if result, err := (domainservice.RubricService{}).Remove(rubric, 1, nil); err == nil || result != nil {
		t.Error("Remove must preserve a nonempty rubric")
	}
	if !slices.Equal(rubric.Criteria(), []domain.Criterion{criterion}) {
		t.Error("a rejected removal changed the input rubric")
	}
}

func TestCriterionCanChangeOnlyAfterAllItsEvaluationsAreCancelled(t *testing.T) {
	criterion := newEvaluationTestCriterion(t, 1, 5)
	rubric := testRubric(t, criterion, newEvaluationTestCriterion(t, 2, 10))
	first := newEvaluationServiceTestStudent(t, 1)
	second := newEvaluationServiceTestStudent(t, 2)
	current := []domain.Evaluation{testEvaluation(t, 1, criterion, 3), testEvaluation(t, 2, criterion, 4)}
	current, err := (domainservice.EvaluationService{}).Cancel(first, rubric, current, 1)
	if err != nil {
		t.Fatal(err)
	}
	service := domainservice.RubricService{}
	replacement := newWeightedTestCriterion(t, 1, 12, 5)
	if _, err := service.Update(rubric, replacement, current); err == nil {
		t.Error("another student's evaluation must still prevent updates")
	}
	current, err = (domainservice.EvaluationService{}).Cancel(second, rubric, current, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Update(rubric, replacement, current); err != nil {
		t.Fatalf("Update after cancelling all evaluations: %v", err)
	}
	if _, err := service.Remove(rubric, 1, current); err != nil {
		t.Fatalf("Remove after cancelling all evaluations: %v", err)
	}
}
