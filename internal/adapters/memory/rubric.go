package memory

import (
	"fmt"

	"pos/internal/application"
	"pos/internal/domain"
)

type RubricRepository struct {
	rubric *domain.Rubric
}

var _ application.RubricRepository = (*RubricRepository)(nil)

func NewRubricRepository() *RubricRepository {
	return &RubricRepository{}
}

func (r *RubricRepository) Get() (*domain.Rubric, error) {
	if r.rubric == nil {
		return nil, nil
	}
	return copyRubric(r.rubric)
}

func (r *RubricRepository) Save(rubric *domain.Rubric) error {
	if rubric == nil {
		return fmt.Errorf("невозможно сохранить отсутствующий рубрикатор")
	}
	copy, err := copyRubric(rubric)
	if err != nil {
		return err
	}
	r.rubric = copy
	return nil
}

func copyRubric(rubric *domain.Rubric) (*domain.Rubric, error) {
	builder := domain.NewRubricBuilder()
	for _, criterion := range rubric.Criteria() {
		if err := builder.AddCriterion(&criterion); err != nil {
			return nil, err
		}
	}
	return builder.Build()
}
