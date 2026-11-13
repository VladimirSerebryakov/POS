package application

type ReportRow struct {
	Student string
	Total   float64
}

func (s *Service) CalculateTotal(studentID int) (float64, error) {
	student, rubric, current, err := s.evaluationContext(studentID)
	if err != nil {
		return 0, err
	}
	return s.calculator.Calculate(student, rubric, current)
}

func (s *Service) Results() ([]ReportRow, error) {
	rubric, err := s.readyRubric()
	if err != nil {
		return nil, err
	}
	students, err := s.loadStudents()
	if err != nil {
		return nil, err
	}
	current, err := s.loadEvaluations()
	if err != nil {
		return nil, err
	}
	rows := make([]ReportRow, 0, len(students))
	for _, student := range students {
		total, err := s.calculator.Calculate(student, rubric, current)
		if err != nil {
			return nil, err
		}
		rows = append(rows, ReportRow{Student: student.FullName(), Total: total})
	}
	return rows, nil
}
