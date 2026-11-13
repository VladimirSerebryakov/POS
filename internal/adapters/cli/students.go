package cli

func (c *Console) addStudent() error {
	name, err := c.text("ФИО: ")
	if err != nil {
		return err
	}
	group, err := c.text("Группа: ")
	if err != nil {
		return err
	}
	student, err := c.service.AddStudent(name, group)
	if err != nil {
		return err
	}
	return c.printf("Студент добавлен. ID: %d\n", student.ID())
}

func (c *Console) listStudents() error {
	students, err := c.service.ListStudents()
	if err != nil {
		return err
	}
	if len(students) == 0 {
		return c.printf("Студентов пока нет.\n")
	}
	for _, student := range students {
		if err := c.printf("%d | %s | %s\n", student.ID(), student.FullName(), student.Group()); err != nil {
			return err
		}
	}
	return nil
}
