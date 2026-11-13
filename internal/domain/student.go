package domain

import (
	"fmt"
	"strings"
)

type Student struct {
	id       int
	fullName string
	group    string
}

func NewStudent(id int, fullName, group string) (*Student, error) {
	student := &Student{
		id:       id,
		fullName: strings.TrimSpace(fullName),
		group:    strings.TrimSpace(group),
	}
	if err := student.Validate(); err != nil {
		return nil, err
	}
	return student, nil
}

func (s Student) Validate() error {
	if s.id <= 0 {
		return fmt.Errorf("ID студента должен быть положительным: получено %d", s.id)
	}
	if s.fullName == "" {
		return fmt.Errorf("ФИО студента не должно быть пустым")
	}
	if s.group == "" {
		return fmt.Errorf("группа студента не должна быть пустой")
	}
	return nil
}

func (s Student) ID() int {
	return s.id
}

func (s Student) FullName() string {
	return s.fullName
}

func (s Student) Group() string {
	return s.group
}
