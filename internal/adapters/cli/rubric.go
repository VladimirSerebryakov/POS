package cli

import (
	"errors"
	"io"
)

func (c *Console) setupRubric() error {
	initialized, err := c.service.RubricInitialized()
	if err != nil || initialized {
		return err
	}
	if err := c.printf("Сначала соберите рубрикатор (критерии оценивания студентов): добавьте критерии и завершите сборку.\n"); err != nil {
		return err
	}
	for {
		choice, err := c.integer("\n1. Добавить критерий\n2. Показать критерии\n3. Завершить сборку\n0. Выход\nВыбор: ")
		if err != nil {
			return err
		}
		switch choice {
		case 0:
			return io.EOF
		case 1:
			err = c.addCriterion()
		case 2:
			err = c.listCriteria()
		case 3:
			err = c.service.CompleteRubric()
			if err == nil {
				return c.printf("Рубрикатор собран.\n")
			}
		default:
			err = c.printf("Нет такого пункта меню.\n")
		}
		if err != nil {
			if err == io.EOF {
				return err
			}
			var terminal *terminalError
			if errors.As(err, &terminal) {
				return err
			}
			if err := c.printError(err); err != nil {
				return err
			}
		}
	}
}

func (c *Console) criterionValues() (string, int, int, error) {
	name, err := c.text("Название критерия: ")
	if err != nil {
		return "", 0, 0, err
	}
	maximum, err := c.integer("Максимальный балл: ")
	if err != nil {
		return "", 0, 0, err
	}
	weight, err := c.integer("Вес (1–5): ")
	return name, maximum, weight, err
}

func (c *Console) addCriterion() error {
	name, maximum, weight, err := c.criterionValues()
	if err != nil {
		return err
	}
	criterion, err := c.service.AddCriterion(name, maximum, weight)
	if err != nil {
		return err
	}
	return c.printf("Критерий добавлен. ID: %d\n", criterion.ID())
}

func (c *Console) listCriteria() error {
	criteria, err := c.service.ListCriteria()
	if err != nil {
		return err
	}
	if len(criteria) == 0 {
		return c.printf("Критериев пока нет.\n")
	}
	for _, criterion := range criteria {
		if err := c.printf("%d | %s | максимум %d | вес %d\n", criterion.ID(), criterion.Name(), criterion.MaxValue(), criterion.Weight().Value()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Console) updateCriterion() error {
	id, err := c.integer("ID критерия: ")
	if err != nil {
		return err
	}
	name, maximum, weight, err := c.criterionValues()
	if err != nil {
		return err
	}
	if _, err := c.service.UpdateCriterion(id, name, maximum, weight); err != nil {
		return err
	}
	return c.printf("Критерий изменён.\n")
}

func (c *Console) removeCriterion() error {
	id, err := c.integer("ID критерия: ")
	if err != nil {
		return err
	}
	if err := c.service.RemoveCriterion(id); err != nil {
		return err
	}
	return c.printf("Критерий удалён.\n")
}
