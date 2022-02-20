package log

import (
	"fmt"
	"runtime"

	"go.uber.org/zap"
)

func GetCallerField(fields ...zap.Field) []zap.Field {
	fpcs := make([]uintptr, 1)
	// Skip 2 levels to get the caller
	n := runtime.Callers(3, fpcs)
	if n == 0 {
		return fields
	}
	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return fields
	}
	fname, linenum := caller.FileLine(fpcs[0] - 1)
	return append([]zap.Field{
		zap.String("callerName", caller.Name()),
		zap.String("linenum", fmt.Sprintf("%s:%d", fname, linenum)),
	}, fields...)

}
