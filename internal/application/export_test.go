package application_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"

	"pos/internal/application"
)

func TestResultsAndExportStrategies(t *testing.T) {
	jsonExporter := &exporterStub{}
	csvExporter := &exporterStub{}
	exporters := map[application.ReportFormat]application.ReportExporter{
		application.ReportJSON: jsonExporter,
		application.ReportCSV:  csvExporter,
	}
	service, _ := newApplication(t, exporters)
	exporters[application.ReportJSON] = &exporterStub{}
	criterion, err := service.AddCriterion("Задавал вопросы", 3, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.CompleteRubric(); err != nil {
		t.Fatal(err)
	}
	first := addTestStudent(t, service, "Иванов Иван")
	addTestStudent(t, service, "Иванов Иван")
	if _, err := service.AssignEvaluation(first.ID(), criterion.ID(), 1); err != nil {
		t.Fatal(err)
	}
	rows, err := service.Results()
	if err != nil || len(rows) != 2 {
		t.Fatalf("unexpected results: %v, %v", rows, err)
	}
	if rows[0].Student != "Иванов Иван" || rows[1].Student != "Иванов Иван" || rows[1].Total != 0 {
		t.Errorf("all students, including namesakes and unevaluated students, must be included: %v", rows)
	}
	if math.Abs(rows[0].Total-10.0/3) > 1e-12 {
		t.Error("the application rounded the domain result")
	}
	var output bytes.Buffer
	if err := service.ExportResults(application.ReportCSV, &output); err != nil {
		t.Fatal(err)
	}
	if csvExporter.calls != 1 || jsonExporter.calls != 0 || output.String() != "report" {
		t.Error("the selected export strategy was not used")
	}
	if len(csvExporter.rows) != 2 || math.Abs(csvExporter.rows[0].Total-10.0/3) > 1e-12 {
		t.Error("the exporter did not receive the current unrounded report")
	}
	if _, err := service.CorrectEvaluation(first.ID(), criterion.ID(), 3); err != nil {
		t.Fatal(err)
	}
	if err := service.ExportResults(application.ReportJSON, io.Discard); err != nil {
		t.Fatal(err)
	}
	if jsonExporter.calls != 1 || jsonExporter.rows[0].Total != 10 {
		t.Error("JSON export did not use the configured strategy and updated scores")
	}
}

func TestExportErrorsArePreserved(t *testing.T) {
	exporter := &exporterStub{}
	service, _ := newApplication(t, map[application.ReportFormat]application.ReportExporter{application.ReportJSON: exporter})
	completeTestRubric(t, service)
	if err := service.ExportResults("xml", io.Discard); !errors.Is(err, application.ErrUnsupportedReportFormat) {
		t.Errorf("got %v, want ErrUnsupportedReportFormat", err)
	}
	if err := service.ExportResults(application.ReportCSV, io.Discard); !errors.Is(err, application.ErrUnsupportedReportFormat) {
		t.Errorf("got %v, want an unconfigured format error", err)
	}
	if err := service.ExportResults(application.ReportJSON, nil); err == nil {
		t.Error("a nil writer must be rejected")
	}
	failure := errors.New("export failed")
	exporter.err = failure
	if err := service.ExportResults(application.ReportJSON, io.Discard); !errors.Is(err, failure) {
		t.Errorf("got %v, want original export error", err)
	}
}
