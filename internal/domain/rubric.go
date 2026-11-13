package domain

import "fmt"

type Rubric struct {
	criteria []Criterion
}

func (r *Rubric) Criteria() []Criterion {
	b := make([]Criterion, len(r.criteria))
	copy(b, r.criteria)
	return b
}

func (r *Rubric) FindCriterion(id int) (Criterion, error) {
	for _, criterion := range r.criteria {
		if id == criterion.id {
			return criterion, nil
		}
	}
	return Criterion{}, fmt.Errorf("критерий с ID %d не найден", id)
}

func (r *Rubric) AddCriterion(criterion *Criterion) error {
	if criterion == nil {
		return fmt.Errorf("критерий не указан")
	}
	if _, err := r.FindCriterion(criterion.id); err == nil {
		return fmt.Errorf("критерий с ID %d уже существует", criterion.id)
	}
	err := criterion.Validate()
	if err != nil {
		return err
	}
	r.criteria = append(r.criteria, *criterion)
	return nil
}
