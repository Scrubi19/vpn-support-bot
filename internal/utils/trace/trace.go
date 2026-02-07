package trace

import (
	"runtime"
	"strings"
)

func GetFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	if len(strings.Split(frame.Function, "*")) == 1 {
		return frame.Function
	}
	return "(" + strings.Split(frame.Function, "*")[1]
}
