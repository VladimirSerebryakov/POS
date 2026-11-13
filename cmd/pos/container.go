package main

import (
	"fmt"
	"io"

	"pos/internal/adapters/cli"
	"pos/internal/adapters/memory"
	"pos/internal/adapters/report"
	"pos/internal/application"
)

type container struct {
	input       io.Reader
	output      io.Writer
	students    application.StudentRepository
	rubrics     application.RubricRepository
	evaluations application.EvaluationRepository
	exporters   map[application.ReportFormat]application.ReportExporter
	app         *application.Service
	console     *cli.Console
}

func newContainer(input io.Reader, output io.Writer) *container {
	return &container{input: input, output: output}
}

func (c *container) Students() application.StudentRepository {
	if c.students == nil {
		c.students = memory.NewStudentRepository()
	}
	return c.students
}

func (c *container) Rubrics() application.RubricRepository {
	if c.rubrics == nil {
		c.rubrics = memory.NewRubricRepository()
	}
	return c.rubrics
}

func (c *container) Evaluations() application.EvaluationRepository {
	if c.evaluations == nil {
		c.evaluations = memory.NewEvaluationRepository()
	}
	return c.evaluations
}

func (c *container) reportExporters() map[application.ReportFormat]application.ReportExporter {
	if c.exporters == nil {
		c.exporters = map[application.ReportFormat]application.ReportExporter{
			application.ReportJSON: report.JSONExporter{},
			application.ReportCSV:  report.CSVExporter{},
		}
	}
	return c.exporters
}

func (c *container) Application() (*application.Service, error) {
	if c.app != nil {
		return c.app, nil
	}
	app, err := application.NewService(c.Students(), c.Rubrics(), c.Evaluations(), c.reportExporters())
	if err != nil {
		return nil, fmt.Errorf("не удалось собрать прикладной сервис: %w", err)
	}
	c.app = app
	return app, nil
}

func (c *container) Console() (*cli.Console, error) {
	if c.console != nil {
		return c.console, nil
	}
	app, err := c.Application()
	if err != nil {
		return nil, err
	}
	console, err := cli.NewConsole(app, c.input, c.output)
	if err != nil {
		return nil, fmt.Errorf("не удалось собрать консоль: %w", err)
	}
	c.console = console
	return console, nil
}
