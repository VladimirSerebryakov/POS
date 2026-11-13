package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFullApplicationAndBothExporters(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	directory := filepath.Join(root, "reports")
	jsonPath := filepath.Join(directory, "report.json")
	csvPath := filepath.Join(directory, "report.csv")
	input := strings.Join([]string{
		"1", "Задавал вопросы", "3", "1", "3",
		"1", "Иванов Иван Иванович", "БПИ-251",
		"7", "1", "1", "1",
		"11", "1",
		"13", "json",
		"13", "csv",
		"0",
	}, "\n")
	var output bytes.Buffer
	if err := run(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"Рубрикатор собран", "3.33", "Работа завершена"} {
		if !strings.Contains(output.String(), fragment) {
			t.Errorf("application output is missing %q", fragment)
		}
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Student string  `json:"student"`
		Total   float64 `json:"total"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Student != "Иванов Иван Иванович" || rows[0].Total != 3.33 {
		t.Fatalf("unexpected JSON report: %v", rows)
	}
	data, err = os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil || len(records) != 2 || records[1][0] != rows[0].Student || records[1][1] != "3.33" {
		t.Fatalf("unexpected CSV report: %v, %v", records, err)
	}
}

type failingOutput struct {
	err error
}

func (w failingOutput) Write([]byte) (int, error) {
	return 0, w.err
}

func TestRunHandlesEOFAndPropagatesStartupAndTerminalErrors(t *testing.T) {
	if err := run(strings.NewReader(""), io.Discard); err != nil {
		t.Fatalf("EOF must end the application normally: %v", err)
	}
	if err := run(nil, io.Discard); err == nil {
		t.Error("a constructor failure must reach the entry point")
	}
	failure := errors.New("output failed")
	if err := run(strings.NewReader("0"), failingOutput{err: failure}); !errors.Is(err, failure) {
		t.Errorf("got %v, want the original terminal error", err)
	}
}
