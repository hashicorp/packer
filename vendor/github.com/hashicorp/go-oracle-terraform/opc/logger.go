package opc

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

const (
	// LogOff turns logging off
	LogOff LogLevelType = 0
	// LogDebug turns logging to debug
	LogDebug LogLevelType = 1
)

// LogLevelType details the constants that log level can be in
type LogLevelType uint

// Logger interface. Should be satisfied by Terraform's logger as well as the Default logger
type Logger interface {
	Log(...interface{})
}

// LoggerFunc details the logger functions
type LoggerFunc func(...interface{})

// Log logs the specified messages
func (f LoggerFunc) Log(args ...interface{}) {
	f(args...)
}

// NewDefaultLogger returns a default logger if one isn't specified during configuration
func NewDefaultLogger() Logger {
	logWriter, err := LogOutput()
	if err != nil {
		log.Fatalf("Error setting up log writer: %s", err)
	}
	return &defaultLogger{
		logger: log.New(logWriter, "", log.LstdFlags),
	}
}

// Default logger to satisfy the logger interface
type defaultLogger struct {
	logger *log.Logger
}

func (l defaultLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}

// LogOutput outputs the requested messages
func LogOutput() (logOutput io.Writer, err error) {
	// Default to nil
	logOutput = ioutil.Discard

	logLevel := LogLevel()
	if logLevel == LogOff {
		return
	}

	// Logging is on, set output to STDERR
	logOutput = os.Stderr
	return
}

// LogLevel gets current Log Level from the ORACLE_LOG env var
func LogLevel() LogLevelType {
	envLevel := os.Getenv("ORACLE_LOG")
	if envLevel == "" {
		return LogOff
	}
	return LogDebug
}
