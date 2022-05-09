// Package logger ...
package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type MessageLevel int

const (
	DebugLevel = MessageLevel(1)
	InfoLevel  = MessageLevel(2)
	WarnLevel  = MessageLevel(3)
	ErrorLevel = MessageLevel(4)
	FatalLevel = MessageLevel(5)
)

type Interface interface {
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	Fatal(message string, args ...interface{})
	Level(level MessageLevel, message string, args ...interface{})
	ErrorIf(err error, msg string) error
	PanicIf(err error, msg string)
}

type Logger struct {
	logger *logrus.Logger
}

func New() *Logger {
	l := &Logger{
		logger: logrus.New(),
	}

	l.logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             true,
		TimestampFormat:           "",
		DisableSorting:            false,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
	},
	)

	return l
}

func (l *Logger) Level(level MessageLevel, message string, args ...interface{}) {
	var lv logrus.Level
	switch level {
	case DebugLevel:
		lv = logrus.DebugLevel
	case InfoLevel:
		lv = logrus.InfoLevel
	case WarnLevel:
		lv = logrus.WarnLevel
	case ErrorLevel:
		lv = logrus.ErrorLevel
	case FatalLevel:
		lv = logrus.FatalLevel
	default:
		lv = logrus.InfoLevel
	}

	if len(args) == 0 {
		l.logger.Log(lv, message)
	} else {
		l.logger.Logf(lv, message, args...)
	}

	if level == FatalLevel {
		os.Exit(1)
	}
}

func (l *Logger) Debug(message string, args ...interface{}) {
	l.Level(DebugLevel, message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.Level(InfoLevel, message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.Level(WarnLevel, message, args...)
}

func (l *Logger) Error(message string, args ...interface{}) {
	l.Level(ErrorLevel, message, args...)
}

func (l *Logger) Fatal(message string, args ...interface{}) {
	l.Level(FatalLevel, message, args...)
}

func (l *Logger) PanicIf(err error, msg string) {
	if err != nil {
		l.Fatal(msg+": %v", err)
	}
}

func (l *Logger) ErrorIf(err error, msg string) error {
	if err != nil {
		l.Error("%s: %v", msg, err)
	}

	return err
}
