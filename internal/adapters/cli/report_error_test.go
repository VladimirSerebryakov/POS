package cli

import (
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestReportErrorLocalizesSystemFailureAndPreservesCauses(t *testing.T) {
	pathError := &os.PathError{Op: "write", Path: "reports/report.json", Err: syscall.ENOSPC}
	err := &reportError{
		operation: "не удалось записать отчёт",
		cause:     errors.Join(pathError, io.ErrClosedPipe),
	}
	if !strings.Contains(err.Error(), "на диске закончилось свободное место") || strings.Contains(err.Error(), syscall.ENOSPC.Error()) {
		t.Fatalf("the system error must be displayed in Russian: %v", err)
	}
	if !errors.Is(err, syscall.ENOSPC) || !errors.Is(err, io.ErrClosedPipe) {
		t.Error("localization must preserve both write and close failures")
	}
	var original *os.PathError
	if !errors.As(err, &original) || original != pathError {
		t.Error("localization must preserve the original filesystem error")
	}
}
