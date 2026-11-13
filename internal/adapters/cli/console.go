package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"pos/internal/application"
)

type Console struct {
	service *application.Service
	input   *bufio.Reader
	output  io.Writer
}

func NewConsole(service *application.Service, input io.Reader, output io.Writer) (*Console, error) {
	if service == nil || input == nil || output == nil {
		return nil, fmt.Errorf("необходимо указать прикладной сервис и потоки ввода и вывода")
	}
	return &Console{service: service, input: bufio.NewReader(input), output: output}, nil
}

func (c *Console) Run() error {
	if err := c.printf("ПОС — преподавательская оценка студентов\n"); err != nil {
		return err
	}
	if err := c.setupRubric(); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	for {
		choice, err := c.integer("\n1. Добавить студента\n2. Список студентов\n3. Критерии\n4. Добавить критерий\n5. Изменить критерий\n6. Удалить критерий\n7. Выставить оценку\n8. Исправить оценку\n9. Отменить оценку\n10. Оценки студента\n11. Итог студента\n12. Итоги всех студентов\n13. Экспорт отчёта\n0. Выход\nВыбор: ")
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if choice == 0 {
			return c.printf("Работа завершена.\n")
		}
		if err := c.execute(choice); err != nil {
			if err == io.EOF {
				return nil
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

func (c *Console) execute(choice int) error {
	switch choice {
	case 1:
		return c.addStudent()
	case 2:
		return c.listStudents()
	case 3:
		return c.listCriteria()
	case 4:
		return c.addCriterion()
	case 5:
		return c.updateCriterion()
	case 6:
		return c.removeCriterion()
	case 7:
		return c.assignEvaluation()
	case 8:
		return c.correctEvaluation()
	case 9:
		return c.cancelEvaluation()
	case 10:
		return c.studentEvaluations()
	case 11:
		return c.studentTotal()
	case 12:
		return c.results()
	case 13:
		return c.exportResults()
	default:
		return c.printf("Нет такого пункта меню.\n")
	}
}
