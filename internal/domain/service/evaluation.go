package service

import (
	"fmt"

	"pos/internal/domain"
)

type EvaluationService struct{}

func (EvaluationService) Assign(student domain.Student, rubric *domain.Rubric, current []domain.Evaluation, criterionID, score int) (*domain.Evaluation, error) {
	criterion, err := evaluationCriterion(student, rubric, criterionID)
	if err != nil {
		return nil, err
	}
	for _, evaluation := range current {
		if evaluation.StudentID() == student.ID() && evaluation.CriterionID() == criterionID {
			return nil, fmt.Errorf("оценка студента с ID %d по критерию с ID %d уже выставлена", student.ID(), criterionID)
		}
	}
	return domain.NewEvaluation(student.ID(), criterion, score)
}

func (EvaluationService) Correct(student domain.Student, rubric *domain.Rubric, current []domain.Evaluation, criterionID, score int) (*domain.Evaluation, error) {
	criterion, err := evaluationCriterion(student, rubric, criterionID)
	if err != nil {
		return nil, err
	}
	index, err := evaluationIndex(current, student.ID(), criterionID)
	if err != nil {
		return nil, err
	}
	return current[index].CorrectScore(criterion, score)
}

func (EvaluationService) Cancel(student domain.Student, rubric *domain.Rubric, current []domain.Evaluation, criterionID int) ([]domain.Evaluation, error) {
	criterion, err := evaluationCriterion(student, rubric, criterionID)
	if err != nil {
		return nil, err
	}
	index, err := evaluationIndex(current, student.ID(), criterionID)
	if err != nil {
		return nil, err
	}
	if err := current[index].Validate(criterion); err != nil {
		return nil, err
	}
	result := make([]domain.Evaluation, 0, len(current)-1)
	result = append(result, current[:index]...)
	result = append(result, current[index+1:]...)
	return result, nil
}

func evaluationCriterion(student domain.Student, rubric *domain.Rubric, criterionID int) (domain.Criterion, error) {
	if err := student.Validate(); err != nil {
		return domain.Criterion{}, fmt.Errorf("некорректные данные оцениваемого студента: %w", err)
	}
	if rubric == nil {
		return domain.Criterion{}, fmt.Errorf("не указан рубрикатор для выставления оценки")
	}
	criterion, err := rubric.FindCriterion(criterionID)
	if err != nil {
		return domain.Criterion{}, fmt.Errorf("некорректный критерий оценки: %w", err)
	}
	return criterion, nil
}

func evaluationIndex(current []domain.Evaluation, studentID, criterionID int) (int, error) {
	index := -1
	for i, evaluation := range current {
		if evaluation.StudentID() == studentID && evaluation.CriterionID() == criterionID {
			if index != -1 {
				return -1, fmt.Errorf("обнаружено несколько оценок студента с ID %d по критерию с ID %d", studentID, criterionID)
			}
			index = i
		}
	}
	if index == -1 {
		return -1, fmt.Errorf("оценка студента с ID %d по критерию с ID %d не найдена", studentID, criterionID)
	}
	return index, nil
}
