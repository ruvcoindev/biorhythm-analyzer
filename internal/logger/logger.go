package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Level представляет уровень логирования
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var levelColors = map[Level]string{
	DEBUG: "\033[36m",
	INFO:  "\033[32m",
	WARN:  "\033[33m",
	ERROR: "\033[31m",
	FATAL: "\033[35m",
}

const resetColor = "\033[0m"

// Logger представляет логгер
type Logger struct {
	level      Level
	infoLog    *log.Logger
	errorLog   *log.Logger
	file       *os.File
	console    bool
	timestamp  bool
	callerInfo bool
}

// Config конфигурация логгера
type Config struct {
	Level      string
	LogFile    string
	Console    bool
	Timestamp  bool
	CallerInfo bool
}

// NewLogger создает новый логгер
func NewLogger(config Config) (*Logger, error) {
	var writers []io.Writer

	level := INFO
	switch strings.ToLower(config.Level) {
	case "debug":
		level = DEBUG
	case "info":
		level = INFO
	case "warn":
		level = WARN
	case "error":
		level = ERROR
	case "fatal":
		level = FATAL
	}

	if config.Console {
		writers = append(writers, os.Stdout)
	}

	var file *os.File
	if config.LogFile != "" {
		logDir := filepath.Dir(config.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		var err error
		file, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writers = append(writers, file)
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	multiWriter := io.MultiWriter(writers...)

	return &Logger{
		level:      level,
		infoLog:    log.New(multiWriter, "", 0),
		errorLog:   log.New(multiWriter, "", 0),
		file:       file,
		console:    config.Console,
		timestamp:  config.Timestamp,
		callerInfo: config.CallerInfo,
	}, nil
}

// Close закрывает файл лога
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

func (l *Logger) formatMessage(level Level, format string, args ...interface{}) string {
	var parts []string

	if l.timestamp {
		parts = append(parts, time.Now().Format("2006-01-02 15:04:05.000"))
	}

	levelName := levelNames[level]
	if l.console {
		parts = append(parts, fmt.Sprintf("%s[%s]%s", levelColors[level], levelName, resetColor))
	} else {
		parts = append(parts, fmt.Sprintf("[%s]", levelName))
	}

	if l.callerInfo {
		if pc, file, line, ok := runtime.Caller(3); ok {
			funcName := runtime.FuncForPC(pc).Name()
			funcName = filepath.Base(funcName)
			fileName := filepath.Base(file)
			parts = append(parts, fmt.Sprintf("%s:%d:%s", fileName, line, funcName))
		}
	}

	message := fmt.Sprintf(format, args...)
	parts = append(parts, message)

	return strings.Join(parts, " ")
}

func (l *Logger) shouldLog(level Level) bool {
	return level >= l.level
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.shouldLog(DEBUG) {
		l.infoLog.Println(l.formatMessage(DEBUG, format, args...))
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.shouldLog(INFO) {
		l.infoLog.Println(l.formatMessage(INFO, format, args...))
	}
}

func (l *Logger) Warn(format string, args ...interface{}) {
	if l.shouldLog(WARN) {
		l.infoLog.Println(l.formatMessage(WARN, format, args...))
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if l.shouldLog(ERROR) {
		l.errorLog.Println(l.formatMessage(ERROR, format, args...))
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	if l.shouldLog(FATAL) {
		l.errorLog.Println(l.formatMessage(FATAL, format, args...))
		os.Exit(1)
	}
}
