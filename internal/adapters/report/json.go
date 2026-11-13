package report

import (
	"encoding/json"
	"fmt"
	"io"

	"pos/internal/application"
)

type JSONExporter struct{}

var _ application.ReportExporter = JSONExporter{}

func (JSONExporter) Export(writer io.Writer, rows []application.ReportRow) error {
	if writer == nil {
		return fmt.Errorf("не указан поток записи отчёта")
	}
	prepared, err := prepareRows(rows)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(prepared)
}
