package application_test

import (
	"errors"
	"testing"

	"pos/internal/application"
)

func TestStudentsGetGeneratedIDsAndAreListedInIDOrder(t *testing.T) {
	service, _ := newApplication(t, nil)
	completeTestRubric(t, service)
	first := addTestStudent(t, service, "Иванов Иван")
	second := addTestStudent(t, service, "Иванов Иван")
	if first.ID() == second.ID() || first.ID() <= 0 || second.ID() <= 0 {
		t.Error("students must get distinct positive IDs")
	}
	students, err := service.ListStudents()
	if err != nil || len(students) != 2 {
		t.Fatalf("unexpected students: %v, %v", students, err)
	}
	if students[0] != first || students[1] != second {
		t.Errorf("unexpected student order: %v", students)
	}
}

func TestInvalidStudentIsNotStored(t *testing.T) {
	service, f := newApplication(t, nil)
	completeTestRubric(t, service)
	if _, err := service.AddStudent(" ", "БПИ-251"); err == nil {
		t.Error("a blank full name must be rejected")
	}
	if _, err := service.AddStudent("Иванов Иван", " "); err == nil {
		t.Error("a blank group must be rejected")
	}
	if len(f.students.data) != 0 {
		t.Error("invalid students were stored")
	}
}

func TestStudentRepositoryErrorsArePreserved(t *testing.T) {
	service, f := newApplication(t, nil)
	completeTestRubric(t, service)
	failure := errors.New("student storage unavailable")
	f.students.idErr = failure
	if _, err := service.AddStudent("Иванов Иван", "БПИ-251"); !errors.Is(err, failure) {
		t.Errorf("got %v, want ID generation error", err)
	}
	f.students.idErr = nil
	f.students.saveErr = failure
	if _, err := service.AddStudent("Иванов Иван", "БПИ-251"); !errors.Is(err, failure) {
		t.Errorf("got %v, want save error", err)
	}
	if len(f.students.data) != 0 {
		t.Error("a failed save added a student")
	}
	f.students.readErr = failure
	if _, err := service.ListStudents(); !errors.Is(err, failure) {
		t.Errorf("got %v, want read error", err)
	}
}

func TestDuplicateGeneratedStudentIDDoesNotOverwriteStudent(t *testing.T) {
	service, f := newApplication(t, nil)
	completeTestRubric(t, service)
	original := addTestStudent(t, service, "Иванов Иван")
	f.students.nextID = original.ID() - 1
	if _, err := service.AddStudent("Петров Пётр", "БПИ-251"); !errors.Is(err, application.ErrStudentAlreadyExists) {
		t.Errorf("got %v, want ErrStudentAlreadyExists", err)
	}
	if len(f.students.data) != 1 || f.students.data[original.ID()] != original {
		t.Error("an ID collision overwrote the existing student")
	}
}
