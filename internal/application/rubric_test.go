package application_test

import (
	"errors"
	"slices"
	"testing"

	"pos/internal/application"
	"pos/internal/domain"
)

func TestRubricLifecycle(t *testing.T) {
	service, f := newApplication(t, nil)
	if initialized, err := service.RubricInitialized(); err != nil || initialized {
		t.Fatalf("a new application must need rubric setup: %v, %v", initialized, err)
	}
	if err := service.CompleteRubric(); err == nil {
		t.Fatal("an empty rubric must not be completed")
	}
	for _, input := range []struct {
		name    string
		maximum int
		weight  int
	}{
		{name: " ", maximum: 10, weight: 3},
		{name: "Не ныл", maximum: 0, weight: 3},
		{name: "Не ныл", maximum: 10, weight: 6},
	} {
		if _, err := service.AddCriterion(input.name, input.maximum, input.weight); err == nil {
			t.Error("AddCriterion must reject invalid values")
		}
	}
	first, err := service.AddCriterion("Ходил на пары", 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	if f.rubrics.data != nil {
		t.Error("a draft must not be saved as a completed rubric")
	}
	if initialized, err := service.RubricInitialized(); err != nil || initialized {
		t.Fatalf("adding a draft criterion must not complete setup: %v, %v", initialized, err)
	}
	draft, err := service.ListCriteria()
	if err != nil || len(draft) != 1 || draft[0] != first {
		t.Fatalf("unexpected draft: %v, %v", draft, err)
	}
	draft[0] = domain.Criterion{}
	if err := service.CompleteRubric(); err != nil {
		t.Fatal(err)
	}
	if initialized, err := service.RubricInitialized(); err != nil || !initialized {
		t.Fatalf("a completed rubric must be ready: %v, %v", initialized, err)
	}
	if err := service.CompleteRubric(); !errors.Is(err, application.ErrRubricAlreadyInitialized) {
		t.Errorf("got %v, want ErrRubricAlreadyInitialized", err)
	}
	second, err := service.AddCriterion("Задавал вопросы", 5, 2)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID() == second.ID() {
		t.Error("criterion IDs must be distinct")
	}
	criteria, err := service.ListCriteria()
	if err != nil || !slices.Equal(criteria, []domain.Criterion{first, second}) {
		t.Fatalf("unexpected completed rubric: %v, %v", criteria, err)
	}
}

func TestCompleteRubricCanRetryFailedSave(t *testing.T) {
	service, f := newApplication(t, nil)
	criterion, err := service.AddCriterion("Ходил на пары", 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("save failed")
	f.rubrics.saveErr = failure
	if err := service.CompleteRubric(); !errors.Is(err, failure) {
		t.Errorf("got %v, want original save error", err)
	}
	if f.rubrics.data != nil {
		t.Error("a failed save must not complete the rubric")
	}
	f.rubrics.saveErr = nil
	if err := service.CompleteRubric(); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(f.rubrics.data.Criteria(), []domain.Criterion{criterion}) {
		t.Error("the draft was lost after a failed save")
	}
}

func TestCriterionChangesRespectEvaluationsOfAllStudents(t *testing.T) {
	service, _ := newApplication(t, nil)
	first, _ := completeTestRubric(t, service)
	addTestStudent(t, service, "Иванов Иван")
	secondStudent := addTestStudent(t, service, "Петров Пётр")
	if _, err := service.AssignEvaluation(secondStudent.ID(), first.ID(), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateCriterion(first.ID(), "Посещал занятия", 20, 5); err == nil {
		t.Error("an evaluation from any student must prevent updates")
	}
	if err := service.RemoveCriterion(first.ID()); err == nil {
		t.Error("an evaluation from any student must prevent removal")
	}
	if err := service.CancelEvaluation(secondStudent.ID(), first.ID()); err != nil {
		t.Fatal(err)
	}
	updated, err := service.UpdateCriterion(first.ID(), "Посещал занятия", 20, 5)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID() != first.ID() || updated.MaxValue() != 20 || updated.Weight().Value() != 5 {
		t.Errorf("unexpected updated criterion: %v", updated)
	}
	if err := service.RemoveCriterion(first.ID()); err != nil {
		t.Fatal(err)
	}
	criteria, err := service.ListCriteria()
	if err != nil || len(criteria) != 1 {
		t.Fatalf("unexpected remaining criteria: %v, %v", criteria, err)
	}
	if err := service.RemoveCriterion(criteria[0].ID()); err == nil {
		t.Error("the final criterion must not be removed")
	}
}

func TestFailedRubricWritesPreserveStoredRubric(t *testing.T) {
	for _, operation := range []string{"add", "update", "remove"} {
		t.Run(operation, func(t *testing.T) {
			service, f := newApplication(t, nil)
			first, _ := completeTestRubric(t, service)
			before := f.rubrics.data.Criteria()
			failure := errors.New("save failed")
			f.rubrics.saveErr = failure
			var err error
			switch operation {
			case "add":
				_, err = service.AddCriterion("Не ныл", 5, 1)
			case "update":
				_, err = service.UpdateCriterion(first.ID(), "Посещал занятия", 20, 5)
			case "remove":
				err = service.RemoveCriterion(first.ID())
			}
			if !errors.Is(err, failure) {
				t.Errorf("got %v, want original save error", err)
			}
			if !slices.Equal(f.rubrics.data.Criteria(), before) {
				t.Error("a failed write changed the stored rubric")
			}
		})
	}
}
