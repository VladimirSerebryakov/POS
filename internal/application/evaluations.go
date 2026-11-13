package application

import (
	"fmt"
	"slices"

	"pos/internal/domain"
)

type CriterionScore struct {
	Criterion domain.Criterion
	Score     *int
}

func (s *Service) AssignEvaluation(studentID, criterionID, score int) (domain.Evaluation, error) {
	student, rubric, current, err := s.evaluationContext(studentID)
	if err != nil {
		return domain.Evaluation{}, err
	}
	evaluation, err := s.evaluator.Assign(student, rubric, current, criterionID, score)
	if err != nil {
		return domain.Evaluation{}, err
	}
	updated := append(slices.Clone(current), *evaluation)
	if err := s.evaluations.Save(updated); err != nil {
		return domain.Evaluation{}, fmt.Errorf("не удалось сохранить оценки: %w", err)
	}
	return *evaluation, nil
}

func (s *Service) CorrectEvaluation(studentID, criterionID, score int) (domain.Evaluation, error) {
	student, rubric, current, err := s.evaluationContext(studentID)
	if err != nil {
		return domain.Evaluation{}, err
	}
	evaluation, err := s.evaluator.Correct(student, rubric, current, criterionID, score)
	if err != nil {
		return domain.Evaluation{}, err
	}
	updated := slices.Clone(current)
	for i, existing := range updated {
		if existing.StudentID() == studentID && existing.CriterionID() == criterionID {
			updated[i] = *evaluation
			break
		}
	}
	if err := s.evaluations.Save(updated); err != nil {
		return domain.Evaluation{}, fmt.Errorf("не удалось сохранить оценки: %w", err)
	}
	return *evaluation, nil
}

func (s *Service) CancelEvaluation(studentID, criterionID int) error {
	student, rubric, current, err := s.evaluationContext(studentID)
	if err != nil {
		return err
	}
	updated, err := s.evaluator.Cancel(student, rubric, current, criterionID)
	if err != nil {
		return err
	}
	if err := s.evaluations.Save(updated); err != nil {
		return fmt.Errorf("не удалось сохранить оценки: %w", err)
	}
	return nil
}

func (s *Service) StudentEvaluations(studentID int) ([]CriterionScore, error) {
	student, rubric, current, err := s.evaluationContext(studentID)
	if err != nil {
		return nil, err
	}
	scores := make(map[int]int)
	for _, evaluation := range current {
		if evaluation.StudentID() == student.ID() {
			scores[evaluation.CriterionID()] = evaluation.Score()
		}
	}
	criteria := rubric.Criteria()
	result := make([]CriterionScore, 0, len(criteria))
	for _, criterion := range criteria {
		row := CriterionScore{Criterion: criterion}
		if score, exists := scores[criterion.ID()]; exists {
			row.Score = &score
		}
		result = append(result, row)
	}
	return result, nil
}
