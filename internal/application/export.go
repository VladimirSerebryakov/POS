package application

import (
	"fmt"
	"io"
)

type ReportFormat string

const (
	ReportJSON ReportFormat = "json"
	ReportCSV  ReportFormat = "csv"
)

func (s *Service) ExportResults(format ReportFormat, writer io.Writer) error {
	exporter, exists := s.exporters[format]
	if !exists {
		return fmt.Errorf("%w: %s", ErrUnsupportedReportFormat, format)
	}
	if writer == nil {
		return fmt.Errorf("не указан поток записи отчёта")
	}
	rows, err := s.Results()
	if err != nil {
		return err
	}
	if err := exporter.Export(writer, rows); err != nil {
		return fmt.Errorf("не удалось выгрузить итоги: %w", err)
	}
	return nil
}
