package cli_test

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pos/internal/adapters/cli"
	"pos/internal/application"
)

func TestConsoleFullScenarioAndReports(t *testing.T) {
	service := newTestApplication(t, nil)
	root := t.TempDir()
	t.Chdir(root)
	directory := filepath.Join(root, "reports")
	if err := os.Mkdir(directory, 0755); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(directory, "report.json")
	csvPath := filepath.Join(directory, "report.csv")
	if err := os.WriteFile(jsonPath, []byte("old report"), 0600); err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{
		"1", "Ходил на пары", "10", "3",
		"1", "Задавал вопросы", "5", "2",
		"2", "3",
		"1", "Иванов Иван Иванович", "БПИ-251",
		"2", "3",
		"7", "1", "1", "0",
		"10", "1",
		"8", "1", "1", "8",
		"7", "1", "2", "5",
		"11", "1", "12",
		"13", "xml", "JSON",
		"13", "csv",
		"0",
	}, "\n")
	output := runConsole(t, service, input)
	for _, fragment := range []string{"Рубрикатор собран", "Иванов Иван Иванович", "оценка: 0", "оценка: не выставлено", "8.80", "Оценка исправлена", "Введите json или csv", "Отчёт сохранён: " + jsonPath, "Отчёт сохранён: " + csvPath, "Работа завершена"} {
		if !strings.Contains(output, fragment) {
			t.Errorf("console output is missing %q", fragment)
		}
	}
	if strings.Contains(output, "Путь к файлу:") {
		t.Error("the export must not prompt for a file path")
	}
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Student string  `json:"student"`
		Total   float64 `json:"total"`
	}
	if err := json.Unmarshal(jsonData, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Student != "Иванов Иван Иванович" || rows[0].Total != 8.8 {
		t.Fatalf("unexpected JSON report: %v", rows)
	}
	csvData, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(strings.NewReader(string(csvData))).ReadAll()
	if err != nil || len(records) != 2 || records[1][0] != rows[0].Student || records[1][1] != "8.80" {
		t.Fatalf("unexpected CSV report: %v, %v", records, err)
	}
	files, err := os.ReadDir(directory)
	if err != nil || len(files) != 2 {
		t.Fatalf("temporary report files were not cleaned up: %v, %v", files, err)
	}
}

func TestConsoleResumesReadyApplicationAndChangesCriteria(t *testing.T) {
	service := newTestApplication(t, nil)
	readyApplication(t, service)
	input := strings.Join([]string{
		"1", "Иванов Иван", "БПИ-251",
		"4", "Не ныл", "5", "1",
		"5", "3", "Участвовал в обсуждениях", "10", "2",
		"3", "6", "3",
		"7", "1", "1", "5",
		"9", "1", "1",
		"0",
	}, "\n")
	output := runConsole(t, service, input)
	if strings.Contains(output, "Сначала соберите рубрикатор") {
		t.Error("a completed rubric must not require setup again")
	}
	for _, fragment := range []string{"Участвовал в обсуждениях", "Критерий удалён", "Оценка отменена"} {
		if !strings.Contains(output, fragment) {
			t.Errorf("console output is missing %q", fragment)
		}
	}
	criteria, err := service.ListCriteria()
	if err != nil || len(criteria) != 2 {
		t.Fatalf("unexpected rubric after removal: %v, %v", criteria, err)
	}
	view, err := service.StudentEvaluations(1)
	if err != nil || view[0].Score != nil {
		t.Fatalf("evaluation cancellation was not saved: %v, %v", view, err)
	}
}

func TestConsoleRecoversFromInvalidInputAndDomainErrors(t *testing.T) {
	service := newTestApplication(t, nil)
	input := strings.Join([]string{
		"3",
		"1", "Не ныл", "десять", "5", "0",
		"1", "Ходил на пары", "10", "3", "3",
		"99",
		"1", "Иванов Иван", "БПИ-251",
		"7", "99", "1", "5",
		"7", "1", "1", "11",
		"7", "1", "1", "5",
		"7", "1", "1", "6",
		"0",
	}, "\n")
	output := runConsole(t, service, input)
	for _, fragment := range []string{
		"Введите целое число",
		"Нет такого пункта меню",
		"Студент с таким ID не найден",
		"рубрикатор должен содержать хотя бы один критерий",
		"вес должен быть целым числом от 1 до 5: получено 0",
		"балл должен быть целым числом от 0 до 10: получено 11",
		"оценка студента с ID 1 по критерию с ID 1 уже выставлена",
		"Работа завершена",
	} {
		if !strings.Contains(output, fragment) {
			t.Errorf("console output is missing %q", fragment)
		}
	}
	if strings.Count(output, "Ошибка:") < 4 {
		t.Error("domain errors were not reported")
	}
	view, err := service.StudentEvaluations(1)
	if err != nil || len(view) != 1 || view[0].Score == nil || *view[0].Score != 5 {
		t.Fatalf("rejected operations changed the valid score: %v, %v", view, err)
	}
}

type brokenReader struct {
	err error
}

func (r brokenReader) Read([]byte) (int, error) {
	return 0, r.err
}

type brokenWriter struct {
	err error
}

func (w brokenWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestConsoleExitsOnEOFIncludingPartialCommands(t *testing.T) {
	for _, input := range []string{"", "0", "1\nХодил на пары", "1\nХодил на пары\n10\n3\n3\n1\nИванов Иван\n"} {
		service := newTestApplication(t, nil)
		runConsole(t, service, input)
	}
}

func TestConsoleReturnsTerminalFailures(t *testing.T) {
	service := newTestApplication(t, nil)
	failure := errors.New("terminal failed")
	for _, streams := range []struct {
		input   io.Reader
		output  io.Writer
		want    error
		message string
	}{
		{input: brokenReader{err: failure}, output: io.Discard, want: failure, message: "не удалось прочитать ввод из консоли"},
		{input: strings.NewReader("0"), output: brokenWriter{err: failure}, want: failure, message: "не удалось вывести сообщение в консоль"},
		{input: strings.NewReader("0"), output: brokenWriter{err: io.EOF}, want: io.EOF, message: "не удалось вывести сообщение в консоль"},
	} {
		console, err := cli.NewConsole(service, streams.input, streams.output)
		if err != nil {
			t.Fatal(err)
		}
		err = console.Run()
		if !errors.Is(err, streams.want) {
			t.Errorf("got %v, want terminal error %v", err, streams.want)
		}
		if err == nil || err.Error() != streams.message {
			t.Errorf("got %v, want a Russian terminal error: %s", err, streams.message)
		}
	}
}

func TestConsoleRejectsMissingDependencies(t *testing.T) {
	service := newTestApplication(t, nil)
	if _, err := cli.NewConsole(nil, strings.NewReader(""), io.Discard); err == nil {
		t.Error("a missing facade must be rejected")
	}
	if _, err := cli.NewConsole(service, nil, io.Discard); err == nil {
		t.Error("a missing input must be rejected")
	}
	if _, err := cli.NewConsole(service, strings.NewReader(""), nil); err == nil {
		t.Error("a missing output must be rejected")
	}
}

type partialExporter struct{}

func (partialExporter) Export(writer io.Writer, rows []application.ReportRow) error {
	if _, err := io.WriteString(writer, "partial report"); err != nil {
		return err
	}
	return io.EOF
}

func TestFailedExportPreservesExistingReportAndContinuesMenu(t *testing.T) {
	service := newTestApplication(t, map[application.ReportFormat]application.ReportExporter{application.ReportJSON: partialExporter{}})
	readyApplication(t, service)
	root := t.TempDir()
	t.Chdir(root)
	directory := filepath.Join(root, "reports")
	if err := os.Mkdir(directory, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "report.json")
	if err := os.WriteFile(path, []byte("previous report"), 0600); err != nil {
		t.Fatal(err)
	}
	output := runConsole(t, service, strings.Join([]string{"13", "json", "0"}, "\n"))
	if !strings.Contains(output, "Ошибка:") || !strings.Contains(output, "Работа завершена") {
		t.Error("an export error must be reported and return to the menu")
	}
	if !strings.Contains(output, "неожиданное завершение потока ввода-вывода") || strings.Contains(output, "EOF") {
		t.Error("the underlying export error must be displayed in Russian")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "previous report" {
		t.Fatalf("a failed export damaged the existing report: %q, %v", data, err)
	}
	files, err := os.ReadDir(directory)
	if err != nil || len(files) != 1 {
		t.Fatalf("failed export left temporary files: %v, %v", files, err)
	}
}

func TestExportReportsDirectoryFailureContinuesMenu(t *testing.T) {
	service := newTestApplication(t, nil)
	readyApplication(t, service)
	root := t.TempDir()
	t.Chdir(root)
	path := filepath.Join(root, "reports")
	if err := os.WriteFile(path, []byte("existing file"), 0600); err != nil {
		t.Fatal(err)
	}
	output := runConsole(t, service, strings.Join([]string{"13", "json", "0"}, "\n"))
	if !strings.Contains(output, "не удалось создать папку отчётов") || !strings.Contains(output, "Работа завершена") {
		t.Error("a reports directory failure must be reported and return to the menu")
	}
	if strings.Contains(output, "Отчёт сохранён:") {
		t.Error("a failed export must not be reported as saved")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "existing file" {
		t.Fatalf("a directory failure changed the existing file: %q, %v", data, err)
	}
}
