package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"pos/internal/application"
)

type CSVExporter struct{}

var _ application.ReportExporter = CSVExporter{}

func (CSVExporter) Export(writer io.Writer, rows []application.ReportRow) error {
	if writer == nil {
		return fmt.Errorf("не указан поток записи отчёта")
	}
	prepared, err := prepareRows(rows)
	if err != nil {
		return err
	}
	encoder := csv.NewWriter(writer)
	if err := encoder.Write([]string{"student", "total"}); err != nil {
		return err
	}
	for _, row := range prepared {
		if err := encoder.Write([]string{row.Student, strconv.FormatFloat(row.Total, 'f', 2, 64)}); err != nil {
			return err
		}
	}
	encoder.Flush()
	return encoder.Error()
}
