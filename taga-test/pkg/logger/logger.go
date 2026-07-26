package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Level defines the logging severity levels.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Logger represents a thread-safe custom logger.
type Logger struct {
	mu     sync.Mutex
	level  Level
	logger *log.Logger
}

var defaultLogger = New(INFO)

// New creates a new logger with the specified level.
func New(level Level) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

// SetLevel updates the default logger's level.
func SetLevel(level Level) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.level = level
}

func (l *Logger) log(level Level, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	prefix := fmt.Sprintf("[%s] [%s] ", timestamp, levelNames[level])
	l.logger.Printf(prefix+format, v...)
}

// Debug logs a debug message.
func Debug(format string, v ...interface{}) {
	defaultLogger.log(DEBUG, format, v...)
}

// Info logs an informational message.
func Info(format string, v ...interface{}) {
	defaultLogger.log(INFO, format, v...)
}

// Warn logs a warning message.
func Warn(format string, v ...interface{}) {
	defaultLogger.log(WARN, format, v...)
}

// Error logs an error message.
func Error(format string, v ...interface{}) {
	defaultLogger.log(ERROR, format, v...)
}
