package domain_test

import (
	"testing"

	"pos/internal/domain"
)

func TestNewStudent(t *testing.T) {
	student, err := domain.NewStudent(1, "  Иванов Иван Иванович \t", "\n БПИ-251 \u00a0")
	if err != nil {
		t.Fatal(err)
	}
	if student.ID() != 1 {
		t.Errorf("ID() = %d, want 1", student.ID())
	}
	if student.FullName() != "Иванов Иван Иванович" {
		t.Errorf("FullName() = %q, want %q", student.FullName(), "Иванов Иван Иванович")
	}
	if student.Group() != "БПИ-251" {
		t.Errorf("Group() = %q, want %q", student.Group(), "БПИ-251")
	}
}

func TestNewStudentRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		fullName string
		group    string
	}{
		{name: "zero ID", id: 0, fullName: "Иванов Иван Иванович", group: "БПИ-251"},
		{name: "negative ID", id: -1, fullName: "Иванов Иван Иванович", group: "БПИ-251"},
		{name: "empty full name", id: 1, fullName: "", group: "БПИ-251"},
		{name: "blank full name", id: 1, fullName: " \t\n\u00a0", group: "БПИ-251"},
		{name: "empty group", id: 1, fullName: "Иванов Иван Иванович", group: ""},
		{name: "blank group", id: 1, fullName: "Иванов Иван Иванович", group: " \t\n\u00a0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			student, err := domain.NewStudent(tt.id, tt.fullName, tt.group)
			if err == nil {
				t.Fatal("NewStudent must reject invalid data")
			}
			if student != nil {
				t.Error("NewStudent must not return a student on error")
			}
		})
	}
}

func TestNewStudentAllowsSameNameForDifferentStudents(t *testing.T) {
	for _, id := range []int{1, 2} {
		if _, err := domain.NewStudent(id, "Иванов Иван Иванович", "БПИ-251"); err != nil {
			t.Fatalf("students with different IDs may have the same full name: %v", err)
		}
	}
}

func TestStudentValidateRejectsZeroValue(t *testing.T) {
	var student domain.Student
	if err := student.Validate(); err == nil {
		t.Error("Validate must reject a zero-value student")
	}
}
