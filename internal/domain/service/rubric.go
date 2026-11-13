package service

import (
	"fmt"

	"pos/internal/domain"
)

type RubricService struct{}

func (RubricService) Update(rubric *domain.Rubric, criterion domain.Criterion, current []domain.Evaluation) (*domain.Rubric, error) {
	if err := criterion.Validate(); err != nil {
		return nil, err
	}
	if err := unusedCriterion(rubric, criterion.ID(), current); err != nil {
		return nil, err
	}
	builder := domain.NewRubricBuilder()
	for _, existing := range rubric.Criteria() {
		if existing.ID() == criterion.ID() {
			existing = criterion
		}
		if err := builder.AddCriterion(&existing); err != nil {
			return nil, err
		}
	}
	return builder.Build()
}

func (RubricService) Remove(rubric *domain.Rubric, criterionID int, current []domain.Evaluation) (*domain.Rubric, error) {
	if err := unusedCriterion(rubric, criterionID, current); err != nil {
		return nil, err
	}
	builder := domain.NewRubricBuilder()
	for _, criterion := range rubric.Criteria() {
		if criterion.ID() == criterionID {
			continue
		}
		if err := builder.AddCriterion(&criterion); err != nil {
			return nil, err
		}
	}
	return builder.Build()
}

func unusedCriterion(rubric *domain.Rubric, criterionID int, current []domain.Evaluation) error {
	if rubric == nil {
		return fmt.Errorf("рубрикатор не указан")
	}
	if _, err := rubric.FindCriterion(criterionID); err != nil {
		return err
	}
	for _, evaluation := range current {
		if evaluation.CriterionID() == criterionID {
			return fmt.Errorf("критерий с ID %d нельзя изменить или удалить: по нему уже выставлены оценки", criterionID)
		}
	}
	return nil
}
