package memory_test

import (
	"errors"
	"math"
	"testing"

	"pos/internal/adapters/memory"
	"pos/internal/application"
)

func TestApplicationWithMemoryRepositories(t *testing.T) {
	students := memory.NewStudentRepository()
	rubrics := memory.NewRubricRepository()
	evaluations := memory.NewEvaluationRepository()
	service, err := application.NewService(students, rubrics, evaluations, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddStudent("Иванов Иван", "БПИ-251"); !errors.Is(err, application.ErrRubricNotInitialized) {
		t.Errorf("got %v, want ErrRubricNotInitialized", err)
	}
	first, err := service.AddCriterion("Ходил на пары", 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.AddCriterion("Задавал вопросы", 5, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.CompleteRubric(); err != nil {
		t.Fatal(err)
	}
	student, err := service.AddStudent("Иванов Иван", "БПИ-251")
	if err != nil {
		t.Fatal(err)
	}
	other, err := service.AddStudent("Петров Пётр", "БПИ-252")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AssignEvaluation(student.ID(), first.ID(), 8); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AssignEvaluation(student.ID(), second.ID(), 5); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AssignEvaluation(other.ID(), first.ID(), 0); err != nil {
		t.Fatal(err)
	}
	checkResult := func(want float64) {
		t.Helper()
		got, err := service.CalculateTotal(student.ID())
		if err != nil || math.Abs(got-want) > 1e-12 {
			t.Fatalf("got result %v, %v; want %v", got, err, want)
		}
	}
	checkResult(8.8)
	if _, err := service.AssignEvaluation(student.ID(), first.ID(), 3); err == nil {
		t.Error("repeated assignment must be rejected")
	}
	if _, err := service.CorrectEvaluation(student.ID(), first.ID(), 11); err == nil {
		t.Error("an invalid correction must be rejected")
	}
	checkResult(8.8)
	if _, err := service.CorrectEvaluation(student.ID(), first.ID(), 5); err != nil {
		t.Fatal(err)
	}
	checkResult(7)
	if err := service.CancelEvaluation(student.ID(), first.ID()); err != nil {
		t.Fatal(err)
	}
	checkResult(4)
	view, err := service.StudentEvaluations(student.ID())
	if err != nil || len(view) != 2 || view[0].Score != nil || view[1].Score == nil || *view[1].Score != 5 {
		t.Fatalf("unexpected evaluation view: %v, %v", view, err)
	}
	if _, err := service.UpdateCriterion(first.ID(), "Посещал занятия", 20, 5); err == nil {
		t.Error("another student's zero score must prevent criterion changes")
	}
	if err := service.RemoveCriterion(first.ID()); err == nil {
		t.Error("another student's zero score must prevent criterion removal")
	}
	if err := service.CancelEvaluation(other.ID(), first.ID()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateCriterion(first.ID(), "Посещал занятия", 20, 5); err != nil {
		t.Fatal(err)
	}
	checkResult(20.0 / 7)
	if err := service.RemoveCriterion(first.ID()); err != nil {
		t.Fatal(err)
	}
	checkResult(10)
	rows, err := service.Results()
	if err != nil || len(rows) != 2 || rows[0].Student != student.FullName() || rows[0].Total != 10 || rows[1].Total != 0 {
		t.Fatalf("unexpected report: %v, %v", rows, err)
	}
	recreated, err := application.NewService(students, rubrics, evaluations, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := recreated.CalculateTotal(student.ID())
	if err != nil || got != 10 {
		t.Fatalf("the application must read state from repositories: %v, %v", got, err)
	}
	third, err := recreated.AddStudent("Сидоров Сидор", "БПИ-253")
	if err != nil || third.ID() <= other.ID() {
		t.Fatalf("student ID generation did not survive facade recreation: %v, %v", third, err)
	}
}
