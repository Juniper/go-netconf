package netconf

import (
	stdlog "log"
)

type LogLevel int

const (
	LogError LogLevel = iota
	LogWarn
	LogInfo
	LogDebug
)

var log Logger = NoopLog{}

type Logger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	Panicf(string, ...interface{})
}

func SetLog(l Logger) {
	log = l
}

type NoopLog struct{}

func (l NoopLog) Debugf(format string, v ...interface{}) {}
func (l NoopLog) Infof(format string, v ...interface{})  {}
func (l NoopLog) Warnf(format string, v ...interface{})  {}
func (l NoopLog) Errorf(format string, v ...interface{}) {}
func (l NoopLog) Fatalf(format string, v ...interface{}) {}
func (l NoopLog) Panicf(format string, v ...interface{}) {}

type StdLog struct {
	level LogLevel
	*stdlog.Logger
}

func NewStdLog(l *stdlog.Logger, level LogLevel) *StdLog {
	return &StdLog{
		Logger: l,
		level:  level,
	}
}

func (l *StdLog) Debugf(format string, v ...interface{}) {
	if l.level >= LogDebug {
		l.Printf(format, v)
	}
}

func (l *StdLog) Infof(format string, v ...interface{}) {
	if l.level >= LogInfo {
		l.Printf(format, v)
	}
}

func (l *StdLog) Warnf(format string, v ...interface{}) {
	if l.level >= LogWarn {
		l.Printf(format, v)
	}
}

func (l *StdLog) Errorf(format string, v ...interface{}) {
	if l.level >= LogError {
		l.Printf(format, v)
	}
}

func (l *StdLog) Fatalf(format string, v ...interface{}) {
	l.Fatalf(format, v)

}

func (l *StdLog) Panicf(format string, v ...interface{}) {
	l.Panicf(format, v)
}
