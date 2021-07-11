package simplelogr

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

var (
	DefaultMessageKey         = "msg"
	DefaultNameKey            = "name"
	DefaultTimestampKey       = "ts"
	DefaultTimestampFormat    = time.RFC3339Nano
	DefaultNameSeparator      = "."
	DefaultTraceVerbosity     = 2
	DefaultDebugVerbosity     = 1
	DefaultSeverityKey        = "severity"
	DefaultErrorKey           = "error"
	DefaultStackTraceKey      = "stacktrace"
	DefaultSeverity           = "INFO"
	DefaultErrorSeverity      = "ERROR"
	DefaultEntrySuffix        = "\n"
	DefaultSpaceSeparator     = " "
	DefaultSeverityThresholds = []SeverityThreshold{
		{Level: DefaultTraceVerbosity, Severity: "TRACE"},
		{Level: DefaultDebugVerbosity, Severity: "DEBUG"},
	}
	DefaultPrimaryColour   = color.New(color.FgHiWhite)
	DefaultSecondaryColour = color.New(color.FgWhite)
	DefaultSeverityColours = map[string]*color.Color{
		"ERROR": color.New(color.FgHiRed),
		"INFO":  color.New(color.FgHiWhite),
		"DEBUG": color.New(color.FgHiBlue),
		"TRACE": color.New(color.FgMagenta),
	}
)

// DefaultTimestampEncoder creates a timestamp encoder using the given formatting string
func DefaultTimestampEncoder(format string) func(t time.Time) string {
	return func(t time.Time) string {
		return t.Format(format)
	}
}

// EncodedError contains information extracted from an error to facilitate logging
type EncodedError struct {
	// Message is the primary message contained in the error, typically the result of error.Error()
	Message string
	// StackTrace is optional stack trace information extracted from the error
	StackTrace string
}

// DefaultErrorEncoder uses an error's error.Error() implementation to populate the EncodedError.Message, and has
// support for github.com/pkg/errors which may have built-in stack traces. If it detects a built-in stack trace it
// will populate the EncodedError.StackTrace with it.
func DefaultErrorEncoder(err error) EncodedError {
	encoded := EncodedError{
		Message: err.Error(),
	}

	type tracedError interface {
		StackTrace() errors.StackTrace
	}
	if traced, ok := err.(tracedError); ok {
		encoded.StackTrace = fmt.Sprintf("%+v", traced.StackTrace())
	}

	return encoded
}

// SeverityThreshold describes a verbosity level at which logs are associated with a given severity level string
type SeverityThreshold struct {
	// Level at which the verbosity level must be greater than or equal to in order to satisfy this threshold
	Level int
	// Severity is the the severity level name
	Severity string
}

// DefaultSeverityEncoder uses the provided defaults and thresholds to convert verbosity levels into a severity name.
// - Errors take precedence, using the provided error severity name.
// - The thresholds are then tested, to identify any specific severities that should be used based on the verbosity level.
// - Finally a default severity is used.
func DefaultSeverityEncoder(defaultSeverity string, errSeverity string, thresholds []SeverityThreshold) func(verbosity int, err error) string {
	return func(level int, err error) string {
		if err != nil {
			return errSeverity
		}

		for _, threshold := range thresholds {
			if level >= threshold.Level {
				return threshold.Severity
			}
		}

		return defaultSeverity
	}
}

// DefaultNameEncoder assembles a list of logger names by joining them using the provided separator.
func DefaultNameEncoder(separator string) func(names []string) string {
	return func(names []string) string {
		return strings.Join(names, separator)
	}
}

// DefaultErrorHandler simply emits logging errors to stderr
func DefaultErrorHandler(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "logging error: %+v", err)
}
