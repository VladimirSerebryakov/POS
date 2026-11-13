package memory

import (
	"fmt"

	"pos/internal/application"
	"pos/internal/domain"
)

type StudentRepository struct {
	students map[int]domain.Student
	lastID   int
}

var _ application.StudentRepository = (*StudentRepository)(nil)

func NewStudentRepository() *StudentRepository {
	return &StudentRepository{students: make(map[int]domain.Student)}
}

func (r *StudentRepository) NextID() (int, error) {
	if r.lastID == int(^uint(0)>>1) {
		return 0, fmt.Errorf("невозможно добавить студента: закончились доступные ID")
	}
	r.lastID++
	return r.lastID, nil
}

func (r *StudentRepository) Add(student domain.Student) error {
	if err := student.Validate(); err != nil {
		return err
	}
	if _, exists := r.students[student.ID()]; exists {
		return fmt.Errorf("%w: %d", application.ErrStudentAlreadyExists, student.ID())
	}
	if r.students == nil {
		r.students = make(map[int]domain.Student)
	}
	r.students[student.ID()] = student
	if student.ID() > r.lastID {
		r.lastID = student.ID()
	}
	return nil
}

func (r *StudentRepository) Find(id int) (domain.Student, error) {
	student, exists := r.students[id]
	if !exists {
		return domain.Student{}, fmt.Errorf("%w: %d", application.ErrStudentNotFound, id)
	}
	return student, nil
}

func (r *StudentRepository) List() ([]domain.Student, error) {
	students := make([]domain.Student, 0, len(r.students))
	for _, student := range r.students {
		students = append(students, student)
	}
	return students, nil
}
