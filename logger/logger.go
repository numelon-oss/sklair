// TAKEN FROM https://github.com/numelon-bespoke/sunc-chan
// Copyright applies. This logger is proprietary software unless both signatories under the bespoke project agreement
// have agreed to make this code publicly available to other projects both within the Bespoke program,
// and to external projects unrelated to Numelon or the second signatory.

// 30/11/2025 Awaiting approval

package logger

import (
	"fmt"
	"io"
	"os"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Cyan   = "\033[36m"
)

type LogLevel uint8

const (
	LevelNone LogLevel = iota
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
)

var levelTags = []struct {
	Raw    string
	Colour string
}{
	{"[NONE]", Reset},
	{"[ERROR]", Red},
	{"[WARN]", Yellow},
	{"[INFO]", Green},
	{"[DEBUG]", Cyan},
}

// Logger is a per-instance logger with level filtering and dual output
type Logger struct {
	level  LogLevel
	stdout io.Writer
}

// New Creates a new logger instance
func New(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		stdout: os.Stdout,
	}
}

func (l *Logger) log(level LogLevel, format string, args ...any) {
	if l.level < level {
		return
	}

	tag := levelTags[level]
	formatted := fmt.Sprintf(format, args...)

	// coloured stdout
	coloured := fmt.Sprintf("%s%s%s | %s\n", tag.Colour, tag.Raw, Reset, formatted)
	_, _ = fmt.Fprint(l.stdout, coloured)
}

func (l *Logger) emptyLine() { _, _ = fmt.Fprintln(l.stdout) }

// shortcut methods
func (l *Logger) Error(format string, args ...any)   { l.log(LevelError, format, args...) }
func (l *Logger) Warning(format string, args ...any) { l.log(LevelWarning, format, args...) }
func (l *Logger) Info(format string, args ...any)    { l.log(LevelInfo, format, args...) }
func (l *Logger) Debug(format string, args ...any)   { l.log(LevelDebug, format, args...) }
func (l *Logger) P(format string, args ...any)       { l.log(LevelNone, format, args...) }
func (l *Logger) EmptyLine()                         { l.emptyLine() }

// shared logger
var shared *Logger

func InitShared(level LogLevel) {
	shared = New(level)
}

// WILL LITERALLY EXPLODE IF SHARED NOT INITIALISED
func Error(format string, args ...any)   { shared.log(LevelError, format, args...) }
func Warning(format string, args ...any) { shared.log(LevelWarning, format, args...) }
func Info(format string, args ...any)    { shared.log(LevelInfo, format, args...) }
func Debug(format string, args ...any)   { shared.log(LevelDebug, format, args...) }
func P(format string, args ...any)       { shared.log(LevelNone, format, args...) }
func EmptyLine()                         { shared.emptyLine() }
