package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"pos/internal/adapters/cli"
	"pos/internal/adapters/memory"
	"pos/internal/adapters/report"
	"pos/internal/application"
)

func newTestApplication(t *testing.T, exporters map[application.ReportFormat]application.ReportExporter) *application.Service {
	t.Helper()
	if exporters == nil {
		exporters = map[application.ReportFormat]application.ReportExporter{
			application.ReportJSON: report.JSONExporter{},
			application.ReportCSV:  report.CSVExporter{},
		}
	}
	service, err := application.NewService(memory.NewStudentRepository(), memory.NewRubricRepository(), memory.NewEvaluationRepository(), exporters)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func readyApplication(t *testing.T, service *application.Service) {
	t.Helper()
	if _, err := service.AddCriterion("Ходил на пары", 10, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddCriterion("Задавал вопросы", 5, 2); err != nil {
		t.Fatal(err)
	}
	if err := service.CompleteRubric(); err != nil {
		t.Fatal(err)
	}
}

func runConsole(t *testing.T, service *application.Service, input string) string {
	t.Helper()
	var output bytes.Buffer
	console, err := cli.NewConsole(service, strings.NewReader(input), &output)
	if err != nil {
		t.Fatal(err)
	}
	if err := console.Run(); err != nil {
		t.Fatal(err)
	}
	return output.String()
}
