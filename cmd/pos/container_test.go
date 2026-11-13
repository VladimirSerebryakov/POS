package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"pos/internal/application"
)

func TestContainerReusesDependenciesAndTheirState(t *testing.T) {
	dependencies := newContainer(strings.NewReader("0"), io.Discard)
	if dependencies.Students() != dependencies.Students() || dependencies.Rubrics() != dependencies.Rubrics() || dependencies.Evaluations() != dependencies.Evaluations() {
		t.Fatal("repositories must be reused within a container")
	}
	app, err := dependencies.Application()
	if err != nil {
		t.Fatal(err)
	}
	again, err := dependencies.Application()
	if err != nil || again != app {
		t.Fatalf("the application must be reused: %v", err)
	}
	criterion, err := app.AddCriterion("Ходил на пары", 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	if err := again.CompleteRubric(); err != nil {
		t.Fatalf("the draft was not preserved: %v", err)
	}
	student, err := app.AddStudent("Иванов Иван Иванович", "БПИ-251")
	if err != nil {
		t.Fatal(err)
	}
	stored, err := dependencies.Students().Find(student.ID())
	if err != nil || stored != student {
		t.Fatalf("the facade uses a different student repository: %v, %v", stored, err)
	}
	if _, err := app.AssignEvaluation(student.ID(), criterion.ID(), 5); err != nil {
		t.Fatal(err)
	}
	evaluations, err := dependencies.Evaluations().List()
	if err != nil || len(evaluations) != 1 || evaluations[0].Score() != 5 {
		t.Fatalf("the facade uses a different evaluation repository: %v, %v", evaluations, err)
	}
	console, err := dependencies.Console()
	if err != nil {
		t.Fatal(err)
	}
	sameConsole, err := dependencies.Console()
	if err != nil || sameConsole != console {
		t.Fatalf("the console must be reused: %v", err)
	}
	separate := newContainer(strings.NewReader("0"), io.Discard)
	if _, err := separate.Students().Find(student.ID()); !errors.Is(err, application.ErrStudentNotFound) {
		t.Error("independent containers must not share students")
	}
	rubric, err := separate.Rubrics().Get()
	if err != nil || rubric != nil {
		t.Fatal("independent containers must not share rubric state")
	}
	evaluations, err = separate.Evaluations().List()
	if err != nil || len(evaluations) != 0 {
		t.Fatal("independent containers must not share evaluations")
	}
}

func TestContainerPropagatesConstructorErrorsAndAllowsRetry(t *testing.T) {
	dependencies := newContainer(strings.NewReader("0"), io.Discard)
	dependencies.exporters = map[application.ReportFormat]application.ReportExporter{"xml": nil}
	if app, err := dependencies.Application(); app != nil || !errors.Is(err, application.ErrUnsupportedReportFormat) {
		t.Fatalf("constructor error was not preserved: %v, %v", app, err)
	}
	dependencies.exporters = nil
	if _, err := dependencies.Application(); err != nil {
		t.Fatalf("failed application construction must not be cached: %v", err)
	}
	withoutInput := newContainer(nil, io.Discard)
	if console, err := withoutInput.Console(); console != nil || err == nil {
		t.Fatal("a console with missing input must not be returned")
	}
	withoutInput.input = strings.NewReader("0")
	if _, err := withoutInput.Console(); err != nil {
		t.Fatalf("failed console construction must not be cached: %v", err)
	}
}
