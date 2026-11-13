package report_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"pos/internal/adapters/report"
	"pos/internal/application"
)

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestExportersProduceEquivalentRoundedReports(t *testing.T) {
	rows := []application.ReportRow{
		{Student: "Иванов, Иван \"Иванович\"\nБПИ", Total: 10.0 / 3},
		{Student: "Петров Пётр", Total: 5.125},
		{Student: "Петров Пётр", Total: 0},
	}
	before := slices.Clone(rows)
	var jsonOutput, csvOutput bytes.Buffer
	if err := (report.JSONExporter{}).Export(&jsonOutput, rows); err != nil {
		t.Fatal(err)
	}
	if err := (report.CSVExporter{}).Export(&csvOutput, rows); err != nil {
		t.Fatal(err)
	}
	var jsonRows []struct {
		Student string  `json:"student"`
		Total   float64 `json:"total"`
	}
	if err := json.Unmarshal(jsonOutput.Bytes(), &jsonRows); err != nil {
		t.Fatal(err)
	}
	csvRows, err := csv.NewReader(&csvOutput).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(jsonRows) != 3 || len(csvRows) != 4 || !slices.Equal(csvRows[0], []string{"student", "total"}) {
		t.Fatalf("unexpected report shape: JSON %v, CSV %v", jsonRows, csvRows)
	}
	wantTotals := []float64{3.33, 5.13, 0}
	for i, row := range jsonRows {
		csvTotal, err := strconv.ParseFloat(csvRows[i+1][1], 64)
		if err != nil {
			t.Fatal(err)
		}
		if row.Student != rows[i].Student || csvRows[i+1][0] != row.Student || row.Total != wantTotals[i] || csvTotal != row.Total {
			t.Errorf("row %d differs between formats or expected values", i)
		}
	}
	if csvRows[3][1] != "0.00" {
		t.Error("CSV must display two decimal places")
	}
	if !reflect.DeepEqual(rows, before) {
		t.Error("exporting changed the input report")
	}
}

func TestEmptyReports(t *testing.T) {
	var jsonOutput, csvOutput bytes.Buffer
	if err := (report.JSONExporter{}).Export(&jsonOutput, nil); err != nil {
		t.Fatal(err)
	}
	if jsonOutput.String() != "[]\n" {
		t.Errorf("empty JSON report = %q, want an empty array", jsonOutput.String())
	}
	if err := (report.CSVExporter{}).Export(&csvOutput, nil); err != nil {
		t.Fatal(err)
	}
	if csvOutput.String() != "student,total\n" {
		t.Errorf("empty CSV report = %q, want only a header", csvOutput.String())
	}
}

func TestExportersRejectInvalidTotalsAndPreserveWriterErrors(t *testing.T) {
	failure := errors.New("write failed")
	for _, exporter := range []application.ReportExporter{report.JSONExporter{}, report.CSVExporter{}} {
		if err := exporter.Export(nil, nil); err == nil {
			t.Error("a nil writer must be rejected")
		}
		if err := exporter.Export(failingWriter{err: failure}, []application.ReportRow{{Student: "Иванов Иван", Total: 5}}); !errors.Is(err, failure) {
			t.Errorf("got %v, want original writer error", err)
		}
		for _, total := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			var output bytes.Buffer
			if err := exporter.Export(&output, []application.ReportRow{{Student: "Иванов Иван", Total: total}}); err == nil {
				t.Error("a non-finite total must be rejected")
			}
			if output.Len() != 0 {
				t.Error("an invalid report must not be partially written")
			}
		}
		if err := exporter.Export(io.Discard, nil); err != nil {
			t.Fatal(err)
		}
	}
}
