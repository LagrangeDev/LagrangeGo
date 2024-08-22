package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/log.go

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/utils/log"
)

func (c *QQClient) SetLogger(logger log.Logger) {
	c.logger = logger
}

func (c *QQClient) info(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Info(log.Getcaller(msg), args...)
	}
}

func (c *QQClient) infoln(msgs ...any) {
	if c.logger != nil {
		c.logger.Info(log.Getcaller(fmt.Sprint(msgs...)))
	}
}

func (c *QQClient) warning(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Warning(log.Getcaller(msg), args...)
	}
}

func (c *QQClient) warningln(msgs ...any) {
	if c.logger != nil {
		c.logger.Warning(log.Getcaller(fmt.Sprint(msgs...)))
	}
}

func (c *QQClient) error(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Error(log.Getcaller(msg), args...)
	}
}

func (c *QQClient) errorln(msgs ...any) {
	if c.logger != nil {
		c.logger.Error(log.Getcaller(fmt.Sprint(msgs...)))
	}
}

func (c *QQClient) debug(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Debug(log.Getcaller(msg), args...)
	}
}

func (c *QQClient) debugln(msgs ...any) {
	if c.logger != nil {
		c.logger.Debug(log.Getcaller(fmt.Sprint(msgs...)))
	}
}

func (c *QQClient) dump(msg string, data []byte, args ...any) {
	if c.logger != nil {
		c.logger.Dump(data, log.Getcaller(msg), args...)
	}
}
