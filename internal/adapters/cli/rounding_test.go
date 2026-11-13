package cli_test

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"pos/internal/application"
)

func TestConsoleRoundsHalfCentTotalsConsistently(t *testing.T) {
	for _, test := range []struct {
		name     string
		commands []string
		want     float64
	}{
		{
			name: "single criterion",
			commands: []string{
				"1", "Ходил на пары", "2000", "1", "3",
				"1", "Иван", "БПИ", "7", "1", "1", "1001",
			},
			want: 5.01,
		},
		{
			name: "different maxima and weights",
			commands: []string{
				"1", "Ходил на пары", "5", "1",
				"1", "Задавал вопросы", "20", "3", "3",
				"1", "Иван", "БПИ", "7", "1", "1", "2", "7", "1", "2", "7",
			},
			want: 3.63,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			commands := append(test.commands, "11", "1", "12", "13", "json", "13", "csv", "0")
			output := runConsole(t, newTestApplication(t, nil), strings.Join(commands, "\n"))
			formatted := strconv.FormatFloat(test.want, 'f', 2, 64)
			if !strings.Contains(output, "Итог студента 1: "+formatted) || !strings.Contains(output, "Иван | "+formatted) {
				t.Errorf("console totals must round to %s", formatted)
			}
			data, err := os.ReadFile(filepath.Join(root, "reports", "report.json"))
			if err != nil {
				t.Fatal(err)
			}
			var rows []application.ReportRow
			if err := json.Unmarshal(data, &rows); err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].Total != test.want {
				t.Errorf("JSON report = %v, want total %s", rows, formatted)
			}
			data, err = os.ReadFile(filepath.Join(root, "reports", "report.csv"))
			if err != nil {
				t.Fatal(err)
			}
			records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
			if err != nil || len(records) != 2 || records[1][1] != formatted {
				t.Errorf("CSV report = %v, %v; want total %s", records, err, formatted)
			}
		})
	}
}
