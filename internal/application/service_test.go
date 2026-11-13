package application_test

import (
	"errors"
	"io"
	"testing"

	"pos/internal/application"
)

func TestScenariosRequireCompletedRubric(t *testing.T) {
	service, _ := newApplication(t, map[application.ReportFormat]application.ReportExporter{
		application.ReportJSON: &exporterStub{},
	})
	tests := []struct {
		name string
		run  func() error
	}{
		{name: "add student", run: func() error { _, err := service.AddStudent("Иванов Иван", "БПИ-251"); return err }},
		{name: "list students", run: func() error { _, err := service.ListStudents(); return err }},
		{name: "assign", run: func() error { _, err := service.AssignEvaluation(1, 1, 3); return err }},
		{name: "correct", run: func() error { _, err := service.CorrectEvaluation(1, 1, 3); return err }},
		{name: "cancel", run: func() error { return service.CancelEvaluation(1, 1) }},
		{name: "view evaluations", run: func() error { _, err := service.StudentEvaluations(1); return err }},
		{name: "calculate", run: func() error { _, err := service.CalculateTotal(1); return err }},
		{name: "results", run: func() error { _, err := service.Results(); return err }},
		{name: "update criterion", run: func() error { _, err := service.UpdateCriterion(1, "Не ныл", 5, 1); return err }},
		{name: "remove criterion", run: func() error { return service.RemoveCriterion(1) }},
		{name: "export", run: func() error { return service.ExportResults(application.ReportJSON, io.Discard) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); !errors.Is(err, application.ErrRubricNotInitialized) {
				t.Errorf("got %v, want ErrRubricNotInitialized", err)
			}
		})
	}
}

func TestNewServiceRejectsMissingDependenciesAndInvalidExporters(t *testing.T) {
	_, f := newApplication(t, nil)
	for _, repositories := range []struct {
		students    application.StudentRepository
		rubrics     application.RubricRepository
		evaluations application.EvaluationRepository
	}{
		{students: nil, rubrics: f.rubrics, evaluations: f.evaluations},
		{students: f.students, rubrics: nil, evaluations: f.evaluations},
		{students: f.students, rubrics: f.rubrics, evaluations: nil},
	} {
		if _, err := application.NewService(repositories.students, repositories.rubrics, repositories.evaluations, nil); err == nil {
			t.Error("NewService must reject missing repositories")
		}
	}
	if _, err := application.NewService(f.students, f.rubrics, f.evaluations, map[application.ReportFormat]application.ReportExporter{"xml": &exporterStub{}}); !errors.Is(err, application.ErrUnsupportedReportFormat) {
		t.Errorf("got %v, want ErrUnsupportedReportFormat", err)
	}
	if _, err := application.NewService(f.students, f.rubrics, f.evaluations, map[application.ReportFormat]application.ReportExporter{application.ReportJSON: nil}); err == nil {
		t.Error("NewService must reject a nil exporter")
	}
}
