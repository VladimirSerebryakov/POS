package domain

import "fmt"

type RubricBuilder struct {
	criteria []Criterion
}

func NewRubricBuilder() *RubricBuilder {
	return &RubricBuilder{make([]Criterion, 0)}
}

func (builder *RubricBuilder) Criteria() []Criterion {
	criteria := make([]Criterion, len(builder.criteria))
	copy(criteria, builder.criteria)
	return criteria
}

func (builder *RubricBuilder) AddCriterion(criterion *Criterion) error {
	if criterion == nil {
		return fmt.Errorf("критерий не указан")
	}
	for _, criterionItem := range builder.criteria {
		if criterionItem.id == criterion.id {
			return fmt.Errorf("критерий с ID %d уже существует", criterion.id)
		}
	}
	err := criterion.Validate()
	if err != nil {
		return fmt.Errorf("некорректный критерий: %w", err)
	}
	builder.criteria = append(builder.criteria, *criterion)
	return nil
}

func (builder *RubricBuilder) Build() (*Rubric, error) {
	if len(builder.criteria) == 0 {
		return nil, fmt.Errorf("рубрикатор должен содержать хотя бы один критерий")
	}

	criteria := make([]Criterion, len(builder.criteria))
	copy(criteria, builder.criteria)

	return &Rubric{criteria: criteria}, nil
}
