package cli

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pos/internal/application"
)

func (c *Console) exportResults() error {
	var format application.ReportFormat
	for {
		value, err := c.text("Формат отчёта (json/csv): ")
		if err != nil {
			return err
		}
		format = application.ReportFormat(strings.ToLower(value))
		if format == application.ReportJSON || format == application.ReportCSV {
			break
		}
		if err := c.printf("Введите json или csv.\n"); err != nil {
			return err
		}
	}
	path, err := filepath.Abs(filepath.Join("reports", "report."+string(format)))
	if err != nil {
		return &reportError{operation: "не удалось определить путь отчёта", cause: err}
	}
	if err := saveReport(path, func(writer io.Writer) error {
		return c.service.ExportResults(format, writer)
	}); err != nil {
		return err
	}
	return c.printf("Отчёт сохранён: %s\n", path)
}

func saveReport(path string, export func(io.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return &reportError{operation: "не удалось создать папку отчётов", cause: err}
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".pos-report-*")
	if err != nil {
		return &reportError{operation: "не удалось создать файл отчёта", cause: err}
	}
	defer os.Remove(file.Name())
	exportErr := export(file)
	if err := errors.Join(exportErr, file.Close()); err != nil {
		return &reportError{operation: "не удалось записать отчёт", cause: err}
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return &reportError{operation: "не удалось сохранить отчёт", cause: err}
	}
	return nil
}
