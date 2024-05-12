package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/log.go

import (
	"fmt"
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

func getcaller(msg string) string {
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

func (c *QQClient) SetLogger(logger Logger) {
	c.logger = logger
}

func (c *QQClient) info(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Info(getcaller(msg), args...)
	}
}

func (c *QQClient) infoln(msgs ...any) {
	if c.logger != nil {
		c.logger.Info(getcaller(fmt.Sprint(msgs...)))
	}
}

func (c *QQClient) warning(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Warning(getcaller(msg), args...)
	}
}

func (c *QQClient) warningln(msgs ...any) {
	if c.logger != nil {
		c.logger.Warning(getcaller(fmt.Sprint(msgs...)))
	}
}

func (c *QQClient) error(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Error(getcaller(msg), args...)
	}
}

func (c *QQClient) errorln(msgs ...any) {
	if c.logger != nil {
		c.logger.Error(getcaller(fmt.Sprint(msgs...)))
	}
}

func (c *QQClient) debug(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Debug(getcaller(msg), args...)
	}
}

func (c *QQClient) debugln(msgs ...any) {
	if c.logger != nil {
		c.logger.Debug(getcaller(fmt.Sprint(msgs...)))
	}
}

func (c *QQClient) dump(msg string, data []byte, args ...any) {
	if c.logger != nil {
		c.logger.Dump(data, getcaller(msg), args...)
	}
}
