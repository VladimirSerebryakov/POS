package memory_test

import (
	"errors"
	"testing"

	"pos/internal/adapters/memory"
	"pos/internal/application"
	"pos/internal/domain"
)

func TestStudentRepositoryAssignsUniqueIDs(t *testing.T) {
	repository := memory.NewStudentRepository()
	first, err := repository.NextID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.NextID()
	if err != nil {
		t.Fatal(err)
	}
	if first <= 0 || second <= 0 || first == second {
		t.Fatalf("expected distinct positive IDs, got %d and %d", first, second)
	}
	if err := repository.Add(testStudent(t, 7, "Иванов Иван")); err != nil {
		t.Fatal(err)
	}
	next, err := repository.NextID()
	if err != nil || next <= 7 {
		t.Fatalf("ID generation did not account for an added student: %d, %v", next, err)
	}
}

func TestStudentRepositoryRejectsDuplicateAndInvalidStudents(t *testing.T) {
	var repository memory.StudentRepository
	original := testStudent(t, 1, "Иванов Иван")
	if err := repository.Add(original); err != nil {
		t.Fatal(err)
	}
	if err := repository.Add(testStudent(t, 1, "Петров Пётр")); !errors.Is(err, application.ErrStudentAlreadyExists) {
		t.Errorf("got %v, want ErrStudentAlreadyExists", err)
	}
	if err := repository.Add(domain.Student{}); err == nil {
		t.Error("Add must reject a zero-value student")
	}
	stored, err := repository.Find(1)
	if err != nil || stored != original {
		t.Fatalf("a rejected addition changed the student: %v, %v", stored, err)
	}
	if _, err := repository.Find(99); !errors.Is(err, application.ErrStudentNotFound) {
		t.Errorf("got %v, want ErrStudentNotFound", err)
	}
	students, err := repository.List()
	if err != nil || len(students) != 1 {
		t.Fatalf("unexpected stored students: %v, %v", students, err)
	}
}

func TestStudentRepositoryReturnsIndependentValues(t *testing.T) {
	repository := memory.NewStudentRepository()
	original := testStudent(t, 1, "Иванов Иван")
	if err := repository.Add(original); err != nil {
		t.Fatal(err)
	}
	students, err := repository.List()
	if err != nil || len(students) != 1 {
		t.Fatalf("unexpected students: %v, %v", students, err)
	}
	students[0] = domain.Student{}
	stored, err := repository.Find(1)
	if err != nil || stored != original {
		t.Fatalf("changing the returned list modified storage: %v, %v", stored, err)
	}
	separate := memory.NewStudentRepository()
	if _, err := separate.Find(1); !errors.Is(err, application.ErrStudentNotFound) {
		t.Error("repository instances must not share data")
	}
}

func TestStudentRepositoryDoesNotOverflowIDs(t *testing.T) {
	repository := memory.NewStudentRepository()
	maximum := int(^uint(0) >> 1)
	if err := repository.Add(testStudent(t, maximum, "Иванов Иван")); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if id, err := repository.NextID(); err == nil || id != 0 {
			t.Errorf("exhausted IDs must return an error, got %d, %v", id, err)
		}
	}
}
