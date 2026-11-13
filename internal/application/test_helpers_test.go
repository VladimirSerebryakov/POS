package application_test

import (
	"fmt"
	"io"
	"slices"
	"testing"

	"pos/internal/application"
	"pos/internal/domain"
)

type studentStore struct {
	data    map[int]domain.Student
	nextID  int
	readErr error
	idErr   error
	saveErr error
}

func (s *studentStore) NextID() (int, error) {
	if s.idErr != nil {
		return 0, s.idErr
	}
	s.nextID++
	return s.nextID, nil
}

func (s *studentStore) Add(student domain.Student) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	if _, exists := s.data[student.ID()]; exists {
		return application.ErrStudentAlreadyExists
	}
	s.data[student.ID()] = student
	return nil
}

func (s *studentStore) Find(id int) (domain.Student, error) {
	if s.readErr != nil {
		return domain.Student{}, s.readErr
	}
	student, exists := s.data[id]
	if !exists {
		return domain.Student{}, application.ErrStudentNotFound
	}
	return student, nil
}

func (s *studentStore) List() ([]domain.Student, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}
	students := make([]domain.Student, 0, len(s.data))
	for id := s.nextID; id >= 1; id-- {
		if student, exists := s.data[id]; exists {
			students = append(students, student)
		}
	}
	return students, nil
}

type rubricStore struct {
	data    *domain.Rubric
	readErr error
	saveErr error
}

func (s *rubricStore) Get() (*domain.Rubric, error) {
	return s.data, s.readErr
}

func (s *rubricStore) Save(rubric *domain.Rubric) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data = rubric
	return nil
}

type evaluationStore struct {
	data    []domain.Evaluation
	readErr error
	saveErr error
	writes  int
}

func (s *evaluationStore) List() ([]domain.Evaluation, error) {
	return s.data, s.readErr
}

func (s *evaluationStore) Save(evaluations []domain.Evaluation) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	seen := make(map[[2]int]bool)
	for _, evaluation := range evaluations {
		key := [2]int{evaluation.StudentID(), evaluation.CriterionID()}
		if seen[key] {
			return fmt.Errorf("duplicate evaluation")
		}
		seen[key] = true
	}
	s.data = slices.Clone(evaluations)
	s.writes++
	return nil
}

type exporterStub struct {
	rows  []application.ReportRow
	calls int
	err   error
}

func (e *exporterStub) Export(writer io.Writer, rows []application.ReportRow) error {
	e.calls++
	if e.err != nil {
		return e.err
	}
	e.rows = slices.Clone(rows)
	_, err := io.WriteString(writer, "report")
	return err
}

type fixture struct {
	students    *studentStore
	rubrics     *rubricStore
	evaluations *evaluationStore
}

func newApplication(t *testing.T, exporters map[application.ReportFormat]application.ReportExporter) (*application.Service, *fixture) {
	t.Helper()
	f := &fixture{
		students:    &studentStore{data: make(map[int]domain.Student)},
		rubrics:     &rubricStore{},
		evaluations: &evaluationStore{},
	}
	service, err := application.NewService(f.students, f.rubrics, f.evaluations, exporters)
	if err != nil {
		t.Fatal(err)
	}
	return service, f
}

func completeTestRubric(t *testing.T, service *application.Service) (domain.Criterion, domain.Criterion) {
	t.Helper()
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
	return first, second
}

func addTestStudent(t *testing.T, service *application.Service, name string) domain.Student {
	t.Helper()
	student, err := service.AddStudent(name, "БПИ-251")
	if err != nil {
		t.Fatal(err)
	}
	return student
}
