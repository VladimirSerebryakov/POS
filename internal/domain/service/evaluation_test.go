package service_test

import (
	"testing"

	"pos/internal/domain"
	domainservice "pos/internal/domain/service"
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

func newEvaluationServiceTestStudent(t *testing.T, id int) domain.Student {
	t.Helper()
	student, err := domain.NewStudent(id, "Иванов Иван Иванович", "БПИ-251")
	if err != nil {
		t.Fatal(err)
	}
	return *student
}

func newEvaluationServiceTestRubric(t *testing.T) *domain.Rubric {
	t.Helper()
	builder := domain.NewRubricBuilder()
	for _, criterion := range []domain.Criterion{
		newEvaluationTestCriterion(t, 1, 5),
		newEvaluationTestCriterion(t, 2, 10),
	} {
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

func TestEvaluationServiceAssign(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	rubric := newEvaluationServiceTestRubric(t)
	otherStudent, err := domain.NewEvaluation(2, newEvaluationTestCriterion(t, 1, 5), 3)
	if err != nil {
		t.Fatal(err)
	}
	otherCriterion, err := domain.NewEvaluation(1, newEvaluationTestCriterion(t, 2, 10), 7)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		current []domain.Evaluation
		score   int
	}{
		{name: "first evaluation with zero score", score: 0},
		{name: "maximum score", score: 5},
		{name: "another student has the same criterion", current: []domain.Evaluation{*otherStudent}, score: 3},
		{name: "student has another criterion", current: []domain.Evaluation{*otherCriterion}, score: 3},
		{name: "both kinds of unrelated evaluations", current: []domain.Evaluation{*otherStudent, *otherCriterion}, score: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := append([]domain.Evaluation(nil), tt.current...)
			evaluation, err := (domainservice.EvaluationService{}).Assign(student, rubric, tt.current, 1, tt.score)
			if err != nil {
				t.Fatal(err)
			}
			if evaluation.StudentID() != 1 || evaluation.CriterionID() != 1 || evaluation.Score() != tt.score {
				t.Errorf("unexpected evaluation: got %v, want student 1, criterion 1, score %d", evaluation, tt.score)
			}
			for i, want := range before {
				if tt.current[i] != want {
					t.Error("Assign changed an existing evaluation")
				}
			}
		})
	}
}

func TestEvaluationServiceAssignRejectsInvalidAssignments(t *testing.T) {
	student := newEvaluationServiceTestStudent(t, 1)
	rubric := newEvaluationServiceTestRubric(t)
	existing, err := domain.NewEvaluation(1, newEvaluationTestCriterion(t, 1, 5), 0)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		student     domain.Student
		rubric      *domain.Rubric
		current     []domain.Evaluation
		criterionID int
		score       int
	}{
		{name: "invalid student", student: domain.Student{}, rubric: rubric, criterionID: 1, score: 3},
		{name: "nil rubric", student: student, rubric: nil, criterionID: 1, score: 3},
		{name: "empty rubric", student: student, rubric: &domain.Rubric{}, criterionID: 1, score: 3},
		{name: "unknown criterion", student: student, rubric: rubric, criterionID: 99, score: 3},
		{name: "zero criterion ID", student: student, rubric: rubric, criterionID: 0, score: 3},
		{name: "negative criterion ID", student: student, rubric: rubric, criterionID: -1, score: 3},
		{name: "negative score", student: student, rubric: rubric, criterionID: 1, score: -1},
		{name: "score above selected maximum", student: student, rubric: rubric, criterionID: 1, score: 6},
		{name: "duplicate evaluation with zero score", student: student, rubric: rubric, current: []domain.Evaluation{*existing}, criterionID: 1, score: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluation, err := (domainservice.EvaluationService{}).Assign(tt.student, tt.rubric, tt.current, tt.criterionID, tt.score)
			if err == nil {
				t.Fatal("Assign must reject an invalid assignment")
			}
			if evaluation != nil {
				t.Error("Assign must not return an evaluation on error")
			}
			if len(tt.current) > 0 && tt.current[0] != *existing {
				t.Error("Assign replaced the existing evaluation")
			}
		})
	}
}
