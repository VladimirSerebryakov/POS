package application

import (
	"fmt"

	"pos/internal/domain"
)

func (s *Service) RubricInitialized() (bool, error) {
	rubric, err := s.rubrics.Get()
	if err != nil {
		return false, fmt.Errorf("не удалось загрузить рубрикатор: %w", err)
	}
	return rubric != nil, nil
}

func (s *Service) AddCriterion(name string, maxValue, weight int) (domain.Criterion, error) {
	rubric, err := s.rubrics.Get()
	if err != nil {
		return domain.Criterion{}, fmt.Errorf("не удалось загрузить рубрикатор: %w", err)
	}
	criteria := s.draft.Criteria()
	if rubric != nil {
		criteria = rubric.Criteria()
	}
	id := 1
	for _, criterion := range criteria {
		if criterion.ID() >= id {
			id = criterion.ID() + 1
		}
	}
	criterion, err := createCriterion(id, name, maxValue, weight)
	if err != nil {
		return domain.Criterion{}, err
	}
	if rubric == nil {
		if err := s.draft.AddCriterion(criterion); err != nil {
			return domain.Criterion{}, err
		}
		return *criterion, nil
	}
	builder := domain.NewRubricBuilder()
	for _, existing := range criteria {
		if err := builder.AddCriterion(&existing); err != nil {
			return domain.Criterion{}, err
		}
	}
	if err := builder.AddCriterion(criterion); err != nil {
		return domain.Criterion{}, err
	}
	updated, err := builder.Build()
	if err != nil {
		return domain.Criterion{}, err
	}
	if err := s.rubrics.Save(updated); err != nil {
		return domain.Criterion{}, fmt.Errorf("не удалось сохранить рубрикатор: %w", err)
	}
	return *criterion, nil
}

func (s *Service) CompleteRubric() error {
	rubric, err := s.rubrics.Get()
	if err != nil {
		return fmt.Errorf("не удалось загрузить рубрикатор: %w", err)
	}
	if rubric != nil {
		return ErrRubricAlreadyInitialized
	}
	rubric, err = s.draft.Build()
	if err != nil {
		return err
	}
	if err := s.rubrics.Save(rubric); err != nil {
		return fmt.Errorf("не удалось сохранить рубрикатор: %w", err)
	}
	s.draft = domain.NewRubricBuilder()
	return nil
}

func (s *Service) ListCriteria() ([]domain.Criterion, error) {
	rubric, err := s.rubrics.Get()
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить рубрикатор: %w", err)
	}
	if rubric == nil {
		return s.draft.Criteria(), nil
	}
	return rubric.Criteria(), nil
}

func (s *Service) UpdateCriterion(id int, name string, maxValue, weight int) (domain.Criterion, error) {
	rubric, err := s.readyRubric()
	if err != nil {
		return domain.Criterion{}, err
	}
	criterion, err := createCriterion(id, name, maxValue, weight)
	if err != nil {
		return domain.Criterion{}, err
	}
	current, err := s.loadEvaluations()
	if err != nil {
		return domain.Criterion{}, err
	}
	updated, err := s.rubricRules.Update(rubric, *criterion, current)
	if err != nil {
		return domain.Criterion{}, err
	}
	if err := s.rubrics.Save(updated); err != nil {
		return domain.Criterion{}, fmt.Errorf("не удалось сохранить рубрикатор: %w", err)
	}
	return *criterion, nil
}

func (s *Service) RemoveCriterion(id int) error {
	rubric, err := s.readyRubric()
	if err != nil {
		return err
	}
	current, err := s.loadEvaluations()
	if err != nil {
		return err
	}
	updated, err := s.rubricRules.Remove(rubric, id, current)
	if err != nil {
		return err
	}
	if err := s.rubrics.Save(updated); err != nil {
		return fmt.Errorf("не удалось сохранить рубрикатор: %w", err)
	}
	return nil
}

func createCriterion(id int, name string, maxValue, weightValue int) (*domain.Criterion, error) {
	weight, err := domain.NewWeight(weightValue)
	if err != nil {
		return nil, err
	}
	return domain.NewCriterion(id, name, weight, maxValue)
}
