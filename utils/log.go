package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

const (
	// 定义颜色代码
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorYellow = "\x1b[33m"
	colorGreen  = "\x1b[32m"
	colorBlue   = "\x1b[34m"
	colorWhite  = "\x1b[37m"
)

//var logger *logrus.Logger

var logger = logrus.New()

func init() {
	//logger = logrus.New()
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&ColoredFormatter{})
	logger.SetOutput(colorable.NewColorableStdout())
}

type ColoredFormatter struct{}

func (f *ColoredFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 获取当前时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 根据日志级别设置相应的颜色
	var levelColor string
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = colorBlue
	case logrus.InfoLevel:
		levelColor = colorGreen
	case logrus.WarnLevel:
		levelColor = colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = colorRed
	default:
		levelColor = colorWhite
	}

	// 构建日志格式
	message := fmt.Sprintf("[%s] [%s%s%s] %-*s[%s]: %s\n",
		timestamp, levelColor, strings.ToUpper(entry.Level.String()), colorReset, 7-len(entry.Level.String()), "", entry.Data["prefix"], entry.Message)
	return []byte(message), nil
}

func GetLogger(prefix string) *logrus.Entry {
	return logger.WithField("prefix", prefix)
}
