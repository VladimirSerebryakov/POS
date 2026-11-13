package service

import (
	"fmt"
	"math/big"

	"pos/internal/domain"
)

type Calculator struct{}

func (Calculator) Calculate(student domain.Student, rubric *domain.Rubric, current []domain.Evaluation) (float64, error) {
	if err := student.Validate(); err != nil {
		return 0, fmt.Errorf("некорректные данные студента для расчёта итога: %w", err)
	}
	if rubric == nil {
		return 0, fmt.Errorf("не указан рубрикатор для расчёта итога")
	}
	criteria := rubric.Criteria()
	if len(criteria) == 0 {
		return 0, fmt.Errorf("невозможно рассчитать итог по пустому рубрикатору")
	}
	scores := make(map[int]int)
	for _, evaluation := range current {
		if evaluation.StudentID() != student.ID() {
			continue
		}
		criterion, err := rubric.FindCriterion(evaluation.CriterionID())
		if err != nil {
			return 0, fmt.Errorf("некорректная оценка для расчёта итога: %w", err)
		}
		if err := evaluation.Validate(criterion); err != nil {
			return 0, fmt.Errorf("некорректная оценка для расчёта итога: %w", err)
		}
		if _, exists := scores[criterion.ID()]; exists {
			return 0, fmt.Errorf("обнаружено несколько оценок студента с ID %d по критерию с ID %d", student.ID(), criterion.ID())
		}
		scores[criterion.ID()] = evaluation.Score()
	}
	weightedScore := new(big.Rat)
	var totalWeight int64
	for _, criterion := range criteria {
		weight := int64(criterion.Weight().Value())
		contribution := new(big.Rat).SetFrac64(int64(scores[criterion.ID()]), int64(criterion.MaxValue()))
		contribution.Mul(contribution, big.NewRat(weight, 1))
		weightedScore.Add(weightedScore, contribution)
		totalWeight += weight
	}
	weightedScore.Mul(weightedScore, big.NewRat(10, totalWeight))
	total, _ := weightedScore.Float64()
	return total, nil
}
