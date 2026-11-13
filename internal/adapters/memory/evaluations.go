package memory

import (
	"fmt"
	"slices"

	"pos/internal/application"
	"pos/internal/domain"
)

type EvaluationRepository struct {
	evaluations []domain.Evaluation
}

var _ application.EvaluationRepository = (*EvaluationRepository)(nil)

func NewEvaluationRepository() *EvaluationRepository {
	return &EvaluationRepository{}
}

func (r *EvaluationRepository) List() ([]domain.Evaluation, error) {
	return slices.Clone(r.evaluations), nil
}

func (r *EvaluationRepository) Save(evaluations []domain.Evaluation) error {
	seen := make(map[[2]int]struct{}, len(evaluations))
	for _, evaluation := range evaluations {
		if evaluation.StudentID() <= 0 || evaluation.CriterionID() <= 0 {
			return fmt.Errorf("ID студента и критерия в оценке должны быть положительными")
		}
		key := [2]int{evaluation.StudentID(), evaluation.CriterionID()}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("повторная оценка студента с ID %d по критерию с ID %d", key[0], key[1])
		}
		seen[key] = struct{}{}
	}
	r.evaluations = slices.Clone(evaluations)
	return nil
}
