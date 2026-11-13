package cli

import "pos/internal/adapters/formatting"

func (c *Console) studentTotal() error {
	studentID, err := c.integer("ID студента: ")
	if err != nil {
		return err
	}
	total, err := c.service.CalculateTotal(studentID)
	if err != nil {
		return err
	}
	return c.printf("Итог студента %d: %.2f\n", studentID, formatting.RoundTotal(total))
}

func (c *Console) results() error {
	rows, err := c.service.Results()
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return c.printf("Студентов пока нет.\n")
	}
	for _, row := range rows {
		if err := c.printf("%s | %.2f\n", row.Student, formatting.RoundTotal(row.Total)); err != nil {
			return err
		}
	}
	return nil
}
