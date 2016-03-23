package netconf

import (
	stdlog "log"
)

// LogLevel represents at which level the app should log
type LogLevel int

// Sets the log levels based on the system being connected to
const (
	LogError LogLevel = iota
	LogWarn
	LogInfo
	LogDebug
)

var log Logger = NoopLog{}

// Logger defines different logging levels for use by a logger
type Logger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	Panicf(string, ...interface{})
}

// SetLog sets the logger as the currently selected logger.
func SetLog(l Logger) {
	log = l
}

// NoopLog is for use when you don't want to actually log out
type NoopLog struct{}

// Debugf adds the formatted debug logging function string
func (l NoopLog) Debugf(format string, v ...interface{}) {}

// Infof adds the formatted information logging function string
func (l NoopLog) Infof(format string, v ...interface{}) {}

// Warnf adds the formatted warning logging function string
func (l NoopLog) Warnf(format string, v ...interface{}) {}

// Errorf adds the formatted error logging function string
func (l NoopLog) Errorf(format string, v ...interface{}) {}

// Fatalf adds the formatted fatal logging function string
func (l NoopLog) Fatalf(format string, v ...interface{}) {}

// Panicf adds the formatted panic logging function string
func (l NoopLog) Panicf(format string, v ...interface{}) {}

// StdLog represents the log level and logger for use in logging
type StdLog struct {
	level LogLevel
	*stdlog.Logger
}

// NewStdLog creates a new StdLog instance with the log level and logger provided
func NewStdLog(l *stdlog.Logger, level LogLevel) *StdLog {
	return &StdLog{
		Logger: l,
		level:  level,
	}
}

// Debugf adds the formatted debug logging function string
func (l *StdLog) Debugf(format string, v ...interface{}) {
	if l.level >= LogDebug {
		l.Printf(format, v)
	}
}

// Infof adds the formatted information logging function string
func (l *StdLog) Infof(format string, v ...interface{}) {
	if l.level >= LogInfo {
		l.Printf(format, v)
	}
}

// Warnf adds the formatted warning logging function string
func (l *StdLog) Warnf(format string, v ...interface{}) {
	if l.level >= LogWarn {
		l.Printf(format, v)
	}
}

// Errorf adds the formatted error logging function string
func (l *StdLog) Errorf(format string, v ...interface{}) {
	if l.level >= LogError {
		l.Printf(format, v)
	}
}

// Fatalf adds the formatted fatal logging function string
func (l *StdLog) Fatalf(format string, v ...interface{}) {
	l.Fatalf(format, v)

}

// Panicf adds the formatted panic logging function string
func (l *StdLog) Panicf(format string, v ...interface{}) {
	l.Panicf(format, v)
}
