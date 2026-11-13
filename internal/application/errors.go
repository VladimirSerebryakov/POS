package application

import "errors"

var (
	ErrRubricNotInitialized     = errors.New("сборка рубрикатора ещё не завершена")
	ErrRubricAlreadyInitialized = errors.New("рубрикатор уже собран")
	ErrStudentNotFound          = errors.New("студент не найден")
	ErrStudentAlreadyExists     = errors.New("студент с таким ID уже существует")
	ErrUnsupportedReportFormat  = errors.New("неподдерживаемый формат отчёта")
)
