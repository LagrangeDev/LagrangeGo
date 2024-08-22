package log

import (
	"runtime"
	"strings"
)

type Logger interface {
	Info(format string, args ...any)
	Warning(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
	Dump(dumped []byte, format string, args ...any)
}

func Getcaller(msg string) string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "[unkcal] " + msg
	}
	fp := runtime.FuncForPC(pc)
	sb := strings.Builder{}
	sb.WriteByte('[')
	if fp == nil {
		sb.WriteString(" unkfun]")
		sb.WriteString(msg)
		return sb.String()
	}
	n := fp.Name()
	i := strings.LastIndex(n, "/")
	if i > 0 && i < len(n) {
		n = n[i+1:]
	}
	sb.WriteString(n)
	sb.WriteString("] ")
	sb.WriteString(msg)
	return sb.String()
}
