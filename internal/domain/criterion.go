package domain

import (
	"fmt"
	"strings"
)

type Criterion struct {
	id       int
	name     string
	weight   Weight
	maxValue int
}

func NewCriterion(id int, name string, weight Weight, maxValue int) (*Criterion, error) {
	name = strings.TrimSpace(name)

	criterion := &Criterion{
		id:       id,
		name:     name,
		weight:   weight,
		maxValue: maxValue,
	}
	err := criterion.Validate()
	if err != nil {
		return nil, err
	}
	return criterion, nil
}

func (c *Criterion) Validate() error {
	if c.id <= 0 {
		return fmt.Errorf("ID критерия должен быть положительным: получено %d", c.id)
	}
	if c.name == "" {
		return fmt.Errorf("название критерия не должно быть пустым")
	}
	if c.maxValue <= 0 {
		return fmt.Errorf("максимальный балл критерия должен быть положительным целым числом: получено %d", c.maxValue)
	}
	if err := c.weight.validate(); err != nil {
		return fmt.Errorf("некорректный вес критерия: %w", err)
	}
	return nil
}

func (c Criterion) ID() int {
	return c.id
}

func (c Criterion) Name() string {
	return c.name
}

func (c Criterion) Weight() Weight {
	return c.weight
}

func (c Criterion) MaxValue() int {
	return c.maxValue
}
