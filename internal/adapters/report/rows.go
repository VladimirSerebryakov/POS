package report

import (
	"fmt"
	"math"

	"pos/internal/adapters/formatting"
	"pos/internal/application"
)

type exportRow struct {
	Student string  `json:"student"`
	Total   float64 `json:"total"`
}

func prepareRows(rows []application.ReportRow) ([]exportRow, error) {
	result := make([]exportRow, 0, len(rows))
	for i, row := range rows {
		total := formatting.RoundTotal(row.Total)
		if math.IsNaN(total) || math.IsInf(total, 0) {
			return nil, fmt.Errorf("некорректный итоговый балл в строке отчёта %d", i+1)
		}
		result = append(result, exportRow{Student: row.Student, Total: total})
	}
	return result, nil
}
