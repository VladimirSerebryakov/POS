package application

import (
	"cmp"
	"fmt"
	"slices"

	"pos/internal/domain"
)

func (s *Service) AddStudent(fullName, group string) (domain.Student, error) {
	if _, err := s.readyRubric(); err != nil {
		return domain.Student{}, err
	}
	id, err := s.students.NextID()
	if err != nil {
		return domain.Student{}, fmt.Errorf("не удалось назначить ID студента: %w", err)
	}
	student, err := domain.NewStudent(id, fullName, group)
	if err != nil {
		return domain.Student{}, err
	}
	if err := s.students.Add(*student); err != nil {
		return domain.Student{}, fmt.Errorf("не удалось сохранить студента: %w", err)
	}
	return *student, nil
}

func (s *Service) ListStudents() ([]domain.Student, error) {
	if _, err := s.readyRubric(); err != nil {
		return nil, err
	}
	return s.loadStudents()
}

func (s *Service) loadStudents() ([]domain.Student, error) {
	students, err := s.students.List()
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить список студентов: %w", err)
	}
	students = slices.Clone(students)
	slices.SortFunc(students, func(a, b domain.Student) int {
		return cmp.Compare(a.ID(), b.ID())
	})
	return students, nil
}
