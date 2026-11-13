package domain

import "fmt"

type Evaluation struct {
	studentID   int
	criterionID int
	score       int
}

func NewEvaluation(studentID int, criterion Criterion, score int) (*Evaluation, error) {
	evaluation := &Evaluation{
		studentID:   studentID,
		criterionID: criterion.ID(),
		score:       score,
	}
	if err := evaluation.Validate(criterion); err != nil {
		return nil, err
	}
	return evaluation, nil
}

func (e Evaluation) Validate(criterion Criterion) error {
	if e.studentID <= 0 {
		return fmt.Errorf("ID студента в оценке должен быть положительным: получено %d", e.studentID)
	}
	if err := criterion.Validate(); err != nil {
		return fmt.Errorf("некорректный критерий оценки: %w", err)
	}
	if e.criterionID != criterion.ID() {
		return fmt.Errorf("ID критерия в оценке (%d) не совпадает с ID переданного критерия (%d)", e.criterionID, criterion.ID())
	}
	if e.score < 0 || e.score > criterion.MaxValue() {
		return fmt.Errorf("балл должен быть целым числом от 0 до %d: получено %d", criterion.MaxValue(), e.score)
	}
	return nil
}

func (e Evaluation) CorrectScore(criterion Criterion, score int) (*Evaluation, error) {
	if err := e.Validate(criterion); err != nil {
		return nil, err
	}
	return NewEvaluation(e.studentID, criterion, score)
}

func (e Evaluation) StudentID() int   { return e.studentID }
func (e Evaluation) CriterionID() int { return e.criterionID }
func (e Evaluation) Score() int       { return e.score }
