package domain

import "fmt"

type Weight struct {
	value int
}

func NewWeight(value int) (Weight, error) {
	weight := Weight{value: value}
	if err := weight.validate(); err != nil {
		return Weight{}, err
	}
	return weight, nil
}

func (w Weight) Value() int {
	return w.value
}

func (w Weight) validate() error {
	if w.value < 1 || w.value > 5 {
		return fmt.Errorf("вес должен быть целым числом от 1 до 5: получено %d", w.value)
	}
	return nil
}
