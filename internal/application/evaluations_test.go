package application_test

import (
	"errors"
	"math"
	"slices"
	"testing"

	"pos/internal/application"
)

func TestEvaluationLifecycleAndView(t *testing.T) {
	service, f := newApplication(t, nil)
	first, second := completeTestRubric(t, service)
	student := addTestStudent(t, service, "Иванов Иван")
	other := addTestStudent(t, service, "Петров Пётр")
	view, err := service.StudentEvaluations(student.ID())
	if err != nil || len(view) != 2 || view[0].Score != nil || view[1].Score != nil {
		t.Fatalf("expected two unevaluated criteria: %v, %v", view, err)
	}
	if _, err := service.AssignEvaluation(student.ID(), first.ID(), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AssignEvaluation(other.ID(), second.ID(), 5); err != nil {
		t.Fatal(err)
	}
	view, err = service.StudentEvaluations(student.ID())
	if err != nil || len(view) != 2 || view[0].Score == nil || *view[0].Score != 0 || view[1].Score != nil {
		t.Fatalf("explicit zero must differ from a missing score: %v, %v", view, err)
	}
	if view[0].Criterion.ID() != first.ID() || view[1].Criterion.ID() != second.ID() {
		t.Error("the view must follow rubric order")
	}
	*view[0].Score = 9
	if f.evaluations.data[0].Score() != 0 {
		t.Error("changing the view modified a stored evaluation")
	}
	if _, err := service.AssignEvaluation(student.ID(), first.ID(), 8); err == nil {
		t.Error("duplicate assignment must be rejected")
	}
	if _, err := service.CorrectEvaluation(student.ID(), first.ID(), 8); err != nil {
		t.Fatal(err)
	}
	total, err := service.CalculateTotal(student.ID())
	if err != nil || math.Abs(total-4.8) > 1e-12 {
		t.Fatalf("got total %v, %v; want 4.8", total, err)
	}
	if _, err := service.CorrectEvaluation(student.ID(), first.ID(), 11); err == nil {
		t.Error("invalid correction must be rejected")
	}
	if err := service.CancelEvaluation(student.ID(), first.ID()); err != nil {
		t.Fatal(err)
	}
	view, err = service.StudentEvaluations(student.ID())
	if err != nil || view[0].Score != nil || view[1].Score != nil {
		t.Fatalf("cancelled evaluation must be absent: %v, %v", view, err)
	}
	total, err = service.CalculateTotal(student.ID())
	if err != nil || total != 0 {
		t.Fatalf("got total %v, %v; want 0", total, err)
	}
	if len(f.evaluations.data) != 1 || f.evaluations.data[0].StudentID() != other.ID() {
		t.Error("cancellation modified another student's evaluation")
	}
}

func TestUnknownStudentDoesNotReachEvaluationWrites(t *testing.T) {
	service, f := newApplication(t, nil)
	first, _ := completeTestRubric(t, service)
	tests := []func() error{
		func() error { _, err := service.AssignEvaluation(99, first.ID(), 3); return err },
		func() error { _, err := service.CorrectEvaluation(99, first.ID(), 3); return err },
		func() error { return service.CancelEvaluation(99, first.ID()) },
		func() error { _, err := service.StudentEvaluations(99); return err },
		func() error { _, err := service.CalculateTotal(99); return err },
	}
	for _, run := range tests {
		if err := run(); !errors.Is(err, application.ErrStudentNotFound) {
			t.Errorf("got %v, want ErrStudentNotFound", err)
		}
	}
	if f.evaluations.writes != 0 {
		t.Error("an unknown student caused an evaluation write")
	}
}

func TestRejectedEvaluationOperationsDoNotPersist(t *testing.T) {
	service, f := newApplication(t, nil)
	first, _ := completeTestRubric(t, service)
	student := addTestStudent(t, service, "Иванов Иван")
	if _, err := service.AssignEvaluation(student.ID(), 99, 3); err == nil {
		t.Error("an unknown criterion must be rejected")
	}
	if _, err := service.AssignEvaluation(student.ID(), first.ID(), 11); err == nil {
		t.Error("an excessive score must be rejected")
	}
	if _, err := service.CorrectEvaluation(student.ID(), first.ID(), 3); err == nil {
		t.Error("a missing evaluation cannot be corrected")
	}
	if err := service.CancelEvaluation(student.ID(), first.ID()); err == nil {
		t.Error("a missing evaluation cannot be cancelled")
	}
	if f.evaluations.writes != 0 {
		t.Error("a rejected operation wrote evaluations")
	}
}

func TestFailedEvaluationSavesPreserveStoredValues(t *testing.T) {
	for _, operation := range []string{"assign", "correct", "cancel"} {
		t.Run(operation, func(t *testing.T) {
			service, f := newApplication(t, nil)
			first, second := completeTestRubric(t, service)
			student := addTestStudent(t, service, "Иванов Иван")
			if _, err := service.AssignEvaluation(student.ID(), first.ID(), 8); err != nil {
				t.Fatal(err)
			}
			before := slices.Clone(f.evaluations.data)
			failure := errors.New("evaluation save failed")
			f.evaluations.saveErr = failure
			var err error
			switch operation {
			case "assign":
				_, err = service.AssignEvaluation(student.ID(), second.ID(), 3)
			case "correct":
				_, err = service.CorrectEvaluation(student.ID(), first.ID(), 2)
			case "cancel":
				err = service.CancelEvaluation(student.ID(), first.ID())
			}
			if !errors.Is(err, failure) {
				t.Errorf("got %v, want original save error", err)
			}
			if !slices.Equal(f.evaluations.data, before) {
				t.Error("failed persistence changed stored evaluations")
			}
		})
	}
}

func TestReadErrorsStopScenarios(t *testing.T) {
	service, f := newApplication(t, nil)
	first, _ := completeTestRubric(t, service)
	student := addTestStudent(t, service, "Иванов Иван")
	failure := errors.New("read failed")
	f.evaluations.readErr = failure
	if _, err := service.AssignEvaluation(student.ID(), first.ID(), 3); !errors.Is(err, failure) {
		t.Errorf("got %v, want evaluation read error", err)
	}
	if _, err := service.UpdateCriterion(first.ID(), "Не ныл", 5, 1); !errors.Is(err, failure) {
		t.Errorf("got %v, want evaluation read error", err)
	}
	f.rubrics.readErr = failure
	if _, err := service.AddCriterion("Не ныл", 5, 1); !errors.Is(err, failure) {
		t.Errorf("got %v, want rubric read error", err)
	}
	if f.evaluations.writes != 0 {
		t.Error("a read error caused an evaluation write")
	}
}
