package logging

import (
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	errHeadColor     = color.New(color.BgRed, color.FgWhite, color.Bold)
	errTextColor     = color.New(color.FgRed)
	successHeadColor = color.New(color.BgGreen, color.FgWhite, color.Bold)
	successTextColor = color.New(color.FgGreen)
)

type Logger struct {
	writer io.Writer
}

func (log *Logger) Success(format string, args ...interface{}) {
	_, _ = successHeadColor.Fprint(log.writer, " SUCCESS ")
	_, _ = successTextColor.Fprintf(log.writer, " "+format+"\n", args...)
}

func (log *Logger) Error(format string, args ...interface{}) {
	_, _ = errHeadColor.Fprint(log.writer, "  ERROR  ")
	_, _ = errTextColor.Fprintf(log.writer, " "+format+"\n", args...)
}

func (log *Logger) Fatal(format string, args ...interface{}) {
	log.Error(format, args...)
	os.Exit(1)
}

var defaultLogger = NewLogger(os.Stderr)

func NewLogger(w io.Writer) *Logger {
	return &Logger{writer: w}
}

func Success(format string, args ...interface{}) {
	defaultLogger.Success(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}
