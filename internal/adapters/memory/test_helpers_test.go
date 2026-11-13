package memory_test

import (
	"testing"

	"pos/internal/domain"
)

func testStudent(t *testing.T, id int, name string) domain.Student {
	t.Helper()
	student, err := domain.NewStudent(id, name, "БПИ-251")
	if err != nil {
		t.Fatal(err)
	}
	return *student
}

func testCriterion(t *testing.T, id int) domain.Criterion {
	t.Helper()
	weight, err := domain.NewWeight(3)
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := domain.NewCriterion(id, "Задавал вопросы", weight, 10)
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

func testEvaluation(t *testing.T, studentID, criterionID, score int) domain.Evaluation {
	t.Helper()
	evaluation, err := domain.NewEvaluation(studentID, testCriterion(t, criterionID), score)
	if err != nil {
		t.Fatal(err)
	}
	return *evaluation
}
