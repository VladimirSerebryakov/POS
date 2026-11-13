package service_test

import (
	"math"
	"testing"

	"pos/internal/domain"
	domainservice "pos/internal/domain/service"
)

func newWeightedTestCriterion(t *testing.T, id, maximum, weightValue int) domain.Criterion {
	t.Helper()
	weight, err := domain.NewWeight(weightValue)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := domain.NewCriterion(id, "Ходил на пары", weight, maximum)
	if err != nil {
		t.Fatal(err)
	}
	return *criterion
}

func testRubric(t *testing.T, criteria ...domain.Criterion) *domain.Rubric {
	t.Helper()
	builder := domain.NewRubricBuilder()
	for _, criterion := range criteria {
		if err := builder.AddCriterion(&criterion); err != nil {
			t.Fatal(err)
		}
	}
	rubric, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	return rubric
}

func testEvaluation(t *testing.T, studentID int, criterion domain.Criterion, score int) domain.Evaluation {
	t.Helper()
	evaluation, err := domain.NewEvaluation(studentID, criterion, score)
	if err != nil {
		t.Fatal(err)
	}
	return *evaluation
}

func TestCalculatorCalculate(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	first := newWeightedTestCriterion(t, 1, 10, 3)
	second := newWeightedTestCriterion(t, 2, 5, 2)
	rubric := testRubric(t, first, second)
	tests := []struct {
		name    string
		current []domain.Evaluation
		want    float64
	}{
		{name: "no evaluations", want: 0},
		{name: "missing criterion counts as zero", current: []domain.Evaluation{testEvaluation(t, 1, first, 8)}, want: 4.8},
		{name: "only second criterion evaluated", current: []domain.Evaluation{testEvaluation(t, 1, second, 5)}, want: 4},
		{name: "different maxima and weights", current: []domain.Evaluation{testEvaluation(t, 1, first, 8), testEvaluation(t, 1, second, 5)}, want: 8.8},
		{name: "all maximum scores", current: []domain.Evaluation{testEvaluation(t, 1, first, 10), testEvaluation(t, 1, second, 5)}, want: 10},
		{name: "explicit zero scores", current: []domain.Evaluation{testEvaluation(t, 1, first, 0), testEvaluation(t, 1, second, 0)}, want: 0},
		{name: "other students do not contribute", current: []domain.Evaluation{testEvaluation(t, 1, first, 8), testEvaluation(t, 2, second, 5)}, want: 4.8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (domainservice.Calculator{}).Calculate(student, rubric, tt.current)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-tt.want) > 1e-12 {
				t.Errorf("Calculate() = %.15f, want %.15f", got, tt.want)
			}
		})
	}
}

func TestCalculatorDoesNotRound(t *testing.T) {
	criterion := newWeightedTestCriterion(t, 1, 3, 1)
	got, err := (domainservice.Calculator{}).Calculate(
		newEvaluationServiceTestStudent(t, 1), testRubric(t, criterion),
		[]domain.Evaluation{testEvaluation(t, 1, criterion, 1)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-10.0/3) > 1e-12 {
		t.Errorf("Calculate() = %.15f, want unrounded 10/3", got)
	}
}

func TestCalculatorHandlesLargeMaximumWithoutIntegerOverflow(t *testing.T) {
	maximum := int(^uint(0) >> 1)
	criterion := newWeightedTestCriterion(t, 1, maximum, 5)
	got, err := (domainservice.Calculator{}).Calculate(
		newEvaluationServiceTestStudent(t, 1), testRubric(t, criterion),
		[]domain.Evaluation{testEvaluation(t, 1, criterion, maximum)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != 10 {
		t.Errorf("Calculate() = %v, want 10", got)
	}
}

func TestCalculatorRejectsInvalidState(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	criterion := newWeightedTestCriterion(t, 1, 5, 3)
	rubric := testRubric(t, criterion)
	evaluation := testEvaluation(t, 1, criterion, 0)
	tests := []struct {
		name    string
		student domain.Student
		rubric  *domain.Rubric
		current []domain.Evaluation
	}{
		{name: "invalid student", student: domain.Student{}, rubric: rubric},
		{name: "nil rubric", student: student, rubric: nil},
		{name: "empty rubric", student: student, rubric: &domain.Rubric{}},
		{name: "duplicate evaluations including zero", student: student, rubric: rubric, current: []domain.Evaluation{evaluation, evaluation}},
		{name: "criterion absent from rubric", student: student, rubric: rubric, current: []domain.Evaluation{testEvaluation(t, 1, newWeightedTestCriterion(t, 99, 10, 1), 8)}},
		{name: "score above current maximum", student: student, rubric: rubric, current: []domain.Evaluation{testEvaluation(t, 1, newWeightedTestCriterion(t, 1, 10, 3), 8)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := (domainservice.Calculator{}).Calculate(tt.student, tt.rubric, tt.current); err == nil {
				t.Error("Calculate must reject invalid domain state")
			}
		})
	}
}

func TestCalculatorUsesCurrentRubricAfterCriterionChanges(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	first := newWeightedTestCriterion(t, 1, 5, 3)
	second := newWeightedTestCriterion(t, 2, 10, 3)
	rubric := testRubric(t, first, second)
	current := []domain.Evaluation{testEvaluation(t, 1, first, 5)}
	checkResult := func(want float64) {
		t.Helper()
		got, err := (domainservice.Calculator{}).Calculate(student, rubric, current)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got-want) > 1e-12 {
			t.Errorf("Calculate() = %v, want %v", got, want)
		}
	}
	checkResult(5)
	var err error
	rubric, err = (domainservice.RubricService{}).Update(rubric, newWeightedTestCriterion(t, 2, 20, 1), current)
	if err != nil {
		t.Fatal(err)
	}
	checkResult(7.5)
	rubric, err = (domainservice.RubricService{}).Remove(rubric, 2, current)
	if err != nil {
		t.Fatal(err)
	}
	checkResult(10)
	third := newWeightedTestCriterion(t, 3, 5, 1)
	if err := rubric.AddCriterion(&third); err != nil {
		t.Fatal(err)
	}
	checkResult(7.5)
}
