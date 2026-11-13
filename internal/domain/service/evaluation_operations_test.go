package service_test

import (
	"math"
	"slices"
	"testing"

	"pos/internal/domain"
	domainservice "pos/internal/domain/service"
)

func TestEvaluationServiceCorrect(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	rubric := newEvaluationServiceTestRubric(t)
	criterion := newEvaluationTestCriterion(t, 1, 5)
	original := testEvaluation(t, 1, criterion, 0)
	current := []domain.Evaluation{
		testEvaluation(t, 2, criterion, 4),
		original,
		testEvaluation(t, 1, newEvaluationTestCriterion(t, 2, 10), 7),
	}
	before := slices.Clone(current)
	for _, score := range []int{0, 3, 5} {
		corrected, err := (domainservice.EvaluationService{}).Correct(student, rubric, current, 1, score)
		if err != nil {
			t.Fatal(err)
		}
		if corrected.StudentID() != 1 || corrected.CriterionID() != 1 || corrected.Score() != score {
			t.Errorf("unexpected corrected evaluation: %v", corrected)
		}
	}
	if !slices.Equal(current, before) {
		t.Error("Correct changed the input collection")
	}
}

func TestEvaluationServiceCorrectRejectsInvalidChanges(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	rubric := newEvaluationServiceTestRubric(t)
	evaluation := testEvaluation(t, 1, newEvaluationTestCriterion(t, 1, 5), 3)
	for _, score := range []int{-1, 6} {
		if corrected, err := (domainservice.EvaluationService{}).Correct(student, rubric, []domain.Evaluation{evaluation}, 1, score); err == nil || corrected != nil {
			t.Errorf("Correct must reject score %d", score)
		}
	}
	if corrected, err := (domainservice.EvaluationService{}).Correct(student, rubric, nil, 1, 3); err == nil || corrected != nil {
		t.Error("Correct must not create a missing evaluation")
	}
}

func TestEvaluationServiceCancelPreservesOtherEvaluationsAndCopiesCollection(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	rubric := newEvaluationServiceTestRubric(t)
	criterion := newEvaluationTestCriterion(t, 1, 5)
	otherStudent := testEvaluation(t, 2, criterion, 4)
	target := testEvaluation(t, 1, criterion, 0)
	otherCriterion := testEvaluation(t, 1, newEvaluationTestCriterion(t, 2, 10), 7)
	current := []domain.Evaluation{otherStudent, target, otherCriterion}
	before := slices.Clone(current)
	remaining, err := (domainservice.EvaluationService{}).Cancel(student, rubric, current, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(remaining, []domain.Evaluation{otherStudent, otherCriterion}) {
		t.Errorf("Cancel() = %v, want only unrelated evaluations", remaining)
	}
	remaining[0] = domain.Evaluation{}
	if !slices.Equal(current, before) {
		t.Error("Cancel or changing its result modified the input collection")
	}
}

func TestEvaluationServiceCancelLastEvaluation(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	rubric := newEvaluationServiceTestRubric(t)
	evaluation := testEvaluation(t, 1, newEvaluationTestCriterion(t, 1, 5), 3)
	remaining, err := (domainservice.EvaluationService{}).Cancel(student, rubric, []domain.Evaluation{evaluation}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 {
		t.Errorf("got %d remaining evaluations, want 0", len(remaining))
	}
	if _, err := (domainservice.EvaluationService{}).Cancel(student, rubric, remaining, 1); err == nil {
		t.Error("Cancel must reject a missing evaluation")
	}
}

func TestEvaluationChangesRejectInvalidContext(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	rubric := newEvaluationServiceTestRubric(t)
	evaluation := testEvaluation(t, 1, newEvaluationTestCriterion(t, 1, 5), 3)
	tests := []struct {
		name    string
		student domain.Student
		rubric  *domain.Rubric
		current []domain.Evaluation
		id      int
	}{
		{name: "invalid student", rubric: rubric, id: 1},
		{name: "nil rubric", student: student, id: 1},
		{name: "unknown criterion", student: student, rubric: rubric, id: 99},
		{name: "another student evaluation", student: newEvaluationServiceTestStudent(t, 2), rubric: rubric, current: []domain.Evaluation{evaluation}, id: 1},
		{name: "duplicate evaluations", student: student, rubric: rubric, current: []domain.Evaluation{evaluation, evaluation}, id: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := domainservice.EvaluationService{}
			if _, err := service.Correct(tt.student, tt.rubric, tt.current, tt.id, 3); err == nil {
				t.Error("Correct must reject invalid context")
			}
			if _, err := service.Cancel(tt.student, tt.rubric, tt.current, tt.id); err == nil {
				t.Error("Cancel must reject invalid context")
			}
		})
	}
}

func TestEvaluationLifecycleUpdatesCalculatedResult(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	rubric := newEvaluationServiceTestRubric(t)
	service := domainservice.EvaluationService{}
	calculator := domainservice.Calculator{}
	evaluation, err := service.Assign(student, rubric, nil, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	current := []domain.Evaluation{*evaluation}
	checkResult := func(want float64) {
		t.Helper()
		got, err := calculator.Calculate(student, rubric, current)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got-want) > 1e-12 {
			t.Errorf("Calculate() = %v, want %v", got, want)
		}
	}
	checkResult(5)
	corrected, err := service.Correct(student, rubric, current, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	current[0] = *corrected
	checkResult(2)
	current, err = service.Cancel(student, rubric, current, 1)
	if err != nil {
		t.Fatal(err)
	}
	checkResult(0)
	if _, err := service.Assign(student, rubric, current, 1, 5); err != nil {
		t.Fatalf("a cancelled criterion can be evaluated again: %v", err)
	}
}
