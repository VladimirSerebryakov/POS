package cli

import "strconv"

func (c *Console) evaluationIDs() (int, int, error) {
	studentID, err := c.integer("ID студента: ")
	if err != nil {
		return 0, 0, err
	}
	criterionID, err := c.integer("ID критерия: ")
	return studentID, criterionID, err
}

func (c *Console) assignEvaluation() error {
	studentID, criterionID, err := c.evaluationIDs()
	if err != nil {
		return err
	}
	score, err := c.integer("Балл: ")
	if err != nil {
		return err
	}
	if _, err := c.service.AssignEvaluation(studentID, criterionID, score); err != nil {
		return err
	}
	return c.printf("Оценка выставлена.\n")
}

func (c *Console) correctEvaluation() error {
	studentID, criterionID, err := c.evaluationIDs()
	if err != nil {
		return err
	}
	score, err := c.integer("Новый балл: ")
	if err != nil {
		return err
	}
	if _, err := c.service.CorrectEvaluation(studentID, criterionID, score); err != nil {
		return err
	}
	return c.printf("Оценка исправлена.\n")
}

func (c *Console) cancelEvaluation() error {
	studentID, criterionID, err := c.evaluationIDs()
	if err != nil {
		return err
	}
	if err := c.service.CancelEvaluation(studentID, criterionID); err != nil {
		return err
	}
	return c.printf("Оценка отменена.\n")
}

func (c *Console) studentEvaluations() error {
	studentID, err := c.integer("ID студента: ")
	if err != nil {
		return err
	}
	rows, err := c.service.StudentEvaluations(studentID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		score := "не выставлено"
		if row.Score != nil {
			score = strconv.Itoa(*row.Score)
		}
		if err := c.printf("%d | %s | максимум %d | оценка: %s\n", row.Criterion.ID(), row.Criterion.Name(), row.Criterion.MaxValue(), score); err != nil {
			return err
		}
	}
	return nil
}
