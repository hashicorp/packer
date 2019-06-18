/*
Package log is the log utilities of sdk
*/
package log

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is the interface of SDK
type Logger interface {
	Debug(...interface{})
	Print(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Warning(...interface{})
	Error(...interface{})
	Panic(...interface{})
	Fatal(...interface{})

	Debugf(string, ...interface{})
	Printf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Warningf(string, ...interface{})
	Errorf(string, ...interface{})
	Panicf(string, ...interface{})
	Fatalf(string, ...interface{})

	SetOutput(io.Writer)
	SetFormatter(Formatter)
	SetLevel(Level)
	GetLevel() Level
}

// BasicLogger is the logger (wrapper for logrus)
type BasicLogger struct {
	*logrus.Logger
}

// Level is the log level of logger (wrapper for logrus)
type Level logrus.Level

// Formatter is the formatter of logger (wrapper for logrus)
type Formatter logrus.Formatter

// New will return a logger pointer
func New() *BasicLogger {
	logger := &BasicLogger{logrus.New()}
	logger.Out = os.Stdout
	logger.Level = logrus.Level(DebugLevel)
	logger.Formatter = &logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	}
	return logger
}

// SetOutput sets the logger output.
func (logger *BasicLogger) SetOutput(out io.Writer) {
	logger.Out = out
}

// SetFormatter sets the logger formatter.
func (logger *BasicLogger) SetFormatter(formatter Formatter) {
	logger.Formatter = logrus.Formatter(formatter)
}

// SetLevel sets the logger level.
func (logger *BasicLogger) SetLevel(level Level) {
	logger.Level = logrus.Level(level)
}

// GetLevel returns the logger level.
func (logger *BasicLogger) GetLevel() Level {
	return Level(logger.Level)
}

var (
	PanicLevel = Level(logrus.PanicLevel)
	FatalLevel = Level(logrus.FatalLevel)
	ErrorLevel = Level(logrus.ErrorLevel)
	WarnLevel  = Level(logrus.WarnLevel)
	InfoLevel  = Level(logrus.InfoLevel)
	DebugLevel = Level(logrus.DebugLevel)

	SetLevel     = func(level Level) { logrus.SetLevel(logrus.Level(level)) }
	GetLevel     = func() Level { return Level(logrus.GetLevel()) }
	SetOutput    = logrus.SetOutput
	SetFormatter = logrus.SetFormatter

	WithError = logrus.WithError
	WithField = logrus.WithField

	Debug   = logrus.Debug
	Print   = logrus.Print
	Info    = logrus.Info
	Warn    = logrus.Warn
	Warning = logrus.Warning
	Error   = logrus.Error
	Panic   = logrus.Panic
	Fatal   = logrus.Fatal

	Debugf   = logrus.Debugf
	Printf   = logrus.Printf
	Infof    = logrus.Infof
	Warnf    = logrus.Warnf
	Warningf = logrus.Warningf
	Errorf   = logrus.Errorf
	Panicf   = logrus.Panicf
	Fatalf   = logrus.Fatalf
)

// Init (Deprecated) will init with level and default output (stdout) and formatter (text without color) to global logger
func Init(level Level) {
	logrus.SetLevel(logrus.Level(level))
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}
