package cli

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"pos/internal/application"
)

type terminalError struct {
	message string
	cause   error
}

func (e *terminalError) Error() string {
	return e.message
}

func (e *terminalError) Unwrap() error {
	return e.cause
}

func (c *Console) printf(format string, args ...any) error {
	if _, err := fmt.Fprintf(c.output, format, args...); err != nil {
		return &terminalError{message: "не удалось вывести сообщение в консоль", cause: err}
	}
	return nil
}

func (c *Console) text(prompt string) (string, error) {
	if err := c.printf("%s", prompt); err != nil {
		return "", err
	}
	line, err := c.input.ReadString('\n')
	if errors.Is(err, io.EOF) && len(line) == 0 {
		return "", io.EOF
	}
	if err != nil && !(errors.Is(err, io.EOF) && len(line) > 0) {
		return "", &terminalError{message: "не удалось прочитать ввод из консоли", cause: err}
	}
	return strings.TrimSpace(line), nil
}

func (c *Console) integer(prompt string) (int, error) {
	for {
		line, err := c.text(prompt)
		if err != nil {
			return 0, err
		}
		value, err := strconv.Atoi(line)
		if err == nil {
			return value, nil
		}
		if err := c.printf("Введите целое число.\n"); err != nil {
			return 0, err
		}
	}
}

func (c *Console) printError(err error) error {
	message := err.Error()
	switch {
	case errors.Is(err, application.ErrRubricNotInitialized):
		message = "Сначала завершите сборку рубрикатора."
	case errors.Is(err, application.ErrRubricAlreadyInitialized):
		message = "Рубрикатор уже собран."
	case errors.Is(err, application.ErrStudentNotFound):
		message = "Студент с таким ID не найден."
	case errors.Is(err, application.ErrStudentAlreadyExists):
		message = "Студент с таким ID уже существует."
	case errors.Is(err, application.ErrUnsupportedReportFormat):
		message = "Этот формат отчёта недоступен."
	}
	return c.printf("Ошибка: %s\n", message)
}
