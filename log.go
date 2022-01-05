package ngorm

import (
	baselog "log"
	"os"
	"sync"
)

type LogLevel uint8

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

type logger struct {
	sync.Mutex
	level                              LogLevel
	dins, iins, wins, eins, fins, pins *baselog.Logger
}

var (
	log = &logger{}
)

func SetLogLevel(level LogLevel) {
	log.Lock()
	defer log.Unlock()
	if level <= 5 {
		log.level = level
		return
	}

	log.level = WarnLevel
}

func init() {
	log.dins = baselog.New(os.Stderr, "[D] ", baselog.LstdFlags)
	log.iins = baselog.New(os.Stderr, "[I] ", baselog.LstdFlags)
	log.wins = baselog.New(os.Stderr, "[W] ", baselog.LstdFlags)
	log.eins = baselog.New(os.Stderr, "[E] ", baselog.LstdFlags)
	log.fins = baselog.New(os.Stderr, "[F] ", baselog.LstdFlags)
	log.pins = baselog.New(os.Stderr, "[P] ", baselog.LstdFlags)
	log.level = InfoLevel
}

func (l *logger) Debugf(msg string, args ...interface{}) {
	if l.level <= DebugLevel {
		l.dins.Printf(msg, args...)
	}
}

func (l *logger) Infof(msg string, args ...interface{}) {
	if l.level <= InfoLevel {
		l.iins.Printf(msg, args...)
	}
}
func (l *logger) Warnf(msg string, args ...interface{}) {
	if l.level <= WarnLevel {
		l.wins.Printf(msg, args...)
	}
}

func (l *logger) Errorf(msg string, args ...interface{}) {
	if l.level <= ErrorLevel {
		l.eins.Printf(msg, args...)
	}
}
func (l *logger) Fatalf(msg string, args ...interface{}) {
	if l.level <= FatalLevel {
		l.fins.Fatalf(msg, args...)
	}
}
func (l *logger) Panicf(msg string, args ...interface{}) {
	if l.level <= PanicLevel {
		l.fins.Panicf(msg, args...)
	}
}

func (l *logger) Debug(args ...interface{}) {
	if l.level <= DebugLevel {
		l.dins.Print(args...)
	}
}

func (l *logger) Info(args ...interface{}) {
	if l.level <= InfoLevel {
		l.iins.Print(args...)
	}
}
func (l *logger) Warn(args ...interface{}) {
	if l.level <= WarnLevel {
		l.wins.Print(args...)
	}
}

func (l *logger) Error(args ...interface{}) {
	if l.level <= ErrorLevel {
		l.eins.Print(args...)
	}
}
func (l *logger) Fatal(args ...interface{}) {
	if l.level <= FatalLevel {
		l.fins.Fatal(args...)
	}
}
func (l *logger) Panic(args ...interface{}) {
	if l.level <= PanicLevel {
		l.fins.Panic(args...)
	}
}
