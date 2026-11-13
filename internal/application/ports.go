package application

import (
	"io"

	"pos/internal/domain"
)

type StudentRepository interface {
	NextID() (int, error)
	Add(student domain.Student) error
	Find(id int) (domain.Student, error)
	List() ([]domain.Student, error)
}

type RubricRepository interface {
	Get() (*domain.Rubric, error)
	Save(rubric *domain.Rubric) error
}

type EvaluationRepository interface {
	List() ([]domain.Evaluation, error)
	Save(evaluations []domain.Evaluation) error
}

type ReportExporter interface {
	Export(writer io.Writer, rows []ReportRow) error
}
