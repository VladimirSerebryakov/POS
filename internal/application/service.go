package application

import (
	"fmt"

	"pos/internal/domain"
	domainservice "pos/internal/domain/service"
)

type Service struct {
	students    StudentRepository
	rubrics     RubricRepository
	evaluations EvaluationRepository
	exporters   map[ReportFormat]ReportExporter
	draft       *domain.RubricBuilder
	evaluator   domainservice.EvaluationService
	calculator  domainservice.Calculator
	rubricRules domainservice.RubricService
}

func NewService(students StudentRepository, rubrics RubricRepository, evaluations EvaluationRepository, exporters map[ReportFormat]ReportExporter) (*Service, error) {
	if students == nil || rubrics == nil || evaluations == nil {
		return nil, fmt.Errorf("необходимо указать хранилища студентов, рубрикатора и оценок")
	}
	strategies := make(map[ReportFormat]ReportExporter, len(exporters))
	for format, exporter := range exporters {
		if format != ReportJSON && format != ReportCSV {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedReportFormat, format)
		}
		if exporter == nil {
			return nil, fmt.Errorf("не указан экспортёр отчёта в формате %s", format)
		}
		strategies[format] = exporter
	}
	return &Service{
		students:    students,
		rubrics:     rubrics,
		evaluations: evaluations,
		exporters:   strategies,
		draft:       domain.NewRubricBuilder(),
	}, nil
}

func (s *Service) readyRubric() (*domain.Rubric, error) {
	rubric, err := s.rubrics.Get()
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить рубрикатор: %w", err)
	}
	if rubric == nil {
		return nil, ErrRubricNotInitialized
	}
	return rubric, nil
}

func (s *Service) loadEvaluations() ([]domain.Evaluation, error) {
	current, err := s.evaluations.List()
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить оценки: %w", err)
	}
	return current, nil
}

func (s *Service) evaluationContext(studentID int) (domain.Student, *domain.Rubric, []domain.Evaluation, error) {
	rubric, err := s.readyRubric()
	if err != nil {
		return domain.Student{}, nil, nil, err
	}
	student, err := s.students.Find(studentID)
	if err != nil {
		return domain.Student{}, nil, nil, fmt.Errorf("не удалось загрузить студента с ID %d: %w", studentID, err)
	}
	current, err := s.loadEvaluations()
	if err != nil {
		return domain.Student{}, nil, nil, err
	}
	return student, rubric, current, nil
}
