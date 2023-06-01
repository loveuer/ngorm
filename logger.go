package ngorm

import (
	"github.com/sirupsen/logrus"
)

type logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}

type defaultLogger struct{}

func (defaultLogger) Debug(msg string) {
	logrus.Debug(msg)
}

func (defaultLogger) Info(msg string) {
	logrus.Info(msg)
}

func (defaultLogger) Warn(msg string) {
	logrus.Warn(msg)
}

func (defaultLogger) Error(msg string) {
	logrus.Error(msg)
}

func (defaultLogger) Fatal(msg string) {
	logrus.Fatal(msg)
}

var (
	DefaultLogger = defaultLogger{}
)
