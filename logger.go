package ngorm

import (
	"sync"

	"github.com/loveuer/nf/nft/log"
)

type logger interface {
	Debug(msg string, data ...any)
	Info(msg string, data ...any)
	Warn(msg string, data ...any)
	Error(msg string, data ...any)
	Panic(msg string, data ...any)
	Fatal(msg string, data ...any)
}

type compatLogger struct {
	lock *sync.Mutex
	l    logger
}

func (c *compatLogger) Debug(msg string) {
	c.l.Debug(prefix + msg)
}

func (c *compatLogger) Info(msg string) {
	c.l.Info(prefix + msg)
}

func (c *compatLogger) Warn(msg string) {
	c.l.Warn(prefix + msg)
}

func (c *compatLogger) Error(msg string) {
	c.l.Error(prefix + msg)
}

func (c *compatLogger) Panic(msg string) {
	c.l.Panic(prefix + msg)
}

func (c *compatLogger) Fatal(msg string) {
	c.l.Fatal(prefix + msg)
}

type defaultLogger struct{}

func (defaultLogger) Debug(msg string, data ...any) {
	log.Debug(prefix+msg, data...)
}

func (l defaultLogger) Info(msg string, data ...any) {
	log.Info(prefix+msg, data...)
}

func (l defaultLogger) Warn(msg string, data ...any) {
	log.Warn(prefix+msg, data...)
}

func (l defaultLogger) Error(msg string, data ...any) {
	log.Error(prefix+msg, data...)
}

func (defaultLogger) Panic(msg string, data ...any) {
	log.Panic(prefix+msg, data...)
}

func (defaultLogger) Fatal(msg string, data ...any) {
	log.Fatal(prefix+msg, data...)
}

const (
	prefix = "NGORM | "
)

var clog = &compatLogger{lock: &sync.Mutex{}, l: defaultLogger{}}
