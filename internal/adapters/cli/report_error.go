package cli

import (
	"errors"
	"io"
	"io/fs"
	"syscall"
)

type reportError struct {
	operation string
	cause     error
}

func (e *reportError) Error() string {
	reason := "ошибка ввода-вывода"
	switch {
	case errors.Is(e.cause, fs.ErrPermission):
		reason = "недостаточно прав доступа"
	case errors.Is(e.cause, fs.ErrNotExist):
		reason = "файл или папка не найдены"
	case errors.Is(e.cause, fs.ErrExist):
		reason = "файл или папка уже существуют"
	case errors.Is(e.cause, syscall.ENOTDIR):
		reason = "часть пути указывает на файл вместо папки"
	case errors.Is(e.cause, syscall.EISDIR):
		reason = "путь указывает на папку вместо файла"
	case errors.Is(e.cause, syscall.ENOSPC):
		reason = "на диске закончилось свободное место"
	case errors.Is(e.cause, syscall.EROFS):
		reason = "файловая система доступна только для чтения"
	case errors.Is(e.cause, fs.ErrClosed), errors.Is(e.cause, io.ErrClosedPipe):
		reason = "поток записи закрыт"
	case errors.Is(e.cause, io.ErrShortWrite):
		reason = "данные записаны не полностью"
	case errors.Is(e.cause, io.EOF), errors.Is(e.cause, io.ErrUnexpectedEOF):
		reason = "неожиданное завершение потока ввода-вывода"
	}
	return e.operation + ": " + reason
}

func (e *reportError) Unwrap() error {
	return e.cause
}
