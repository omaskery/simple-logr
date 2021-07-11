package simplelogr

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
)

// JSONLogSink emits structured JSON representations of log Entry objects
type JSONLogSink struct {
	options JSONLogSinkOptions
}

// NewJSONLogSink creates a new JSONLogSink with the provided options
func NewJSONLogSink(options JSONLogSinkOptions) *JSONLogSink {
	return &JSONLogSink{
		options: options,
	}
}

// Log implements LogSink, encoding the given Entry as JSON before writing it to the configured io.Writer
func (j JSONLogSink) Log(e Entry) error {
	obj := map[string]interface{}{}

	if j.options.TimestampKey != "" {
		obj[j.options.TimestampKey] = j.options.TimestampEncoder(e.Timestamp)
	}

	if j.options.SeverityKey != "" {
		obj[j.options.SeverityKey] = j.options.SeverityEncoder(e.Level, e.Error)
	}

	if len(e.Names) > 0 && j.options.NameKey != "" {
		obj[j.options.NameKey] = j.options.NameEncoder(e.Names)
	}

	if e.Message != "" && j.options.MessageKey != "" {
		obj[j.options.MessageKey] = e.Message
	}

	if e.Error != nil && (j.options.ErrorKey != "" || j.options.StackTraceKey != "") {
		encodedErr := j.options.ErrorEncoder(e.Error)
		if j.options.ErrorKey != "" && encodedErr.Message != "" {
			obj[j.options.ErrorKey] = encodedErr.Message
		}
		if j.options.StackTraceKey != "" && encodedErr.StackTrace != "" {
			obj[j.options.StackTraceKey] = encodedErr.StackTrace
		}
	}

	for i := 0; i < len(e.KVs); i += 2 {
		k := e.KVs[i]
		v := e.KVs[i+1]

		kStr, ok := k.(string)
		if !ok {
			return errors.Errorf("logging keys must be strings, got %T: %v", k, k)
		}

		obj[kStr] = v
	}

	if err := json.NewEncoder(j.options.Output).Encode(obj); err != nil {
		return errors.Wrap(err, "failed to encode log entry as JSON")
	}

	return nil
}

// JSONLogSinkOptions configures the behaviour of a JSONLogSink
type JSONLogSinkOptions struct {
	// Output configures where to write structured JSON logs to
	Output io.Writer
	// SeverityKey determines the top level JSON object key to store the log severity name in
	SeverityKey string
	// SeverityEncoder identifies the severity name based on the verbosity level and the presence of any errors
	SeverityEncoder func(level int, err error) string
	// NameKey determines the top level JSON object key to store the logger name in
	NameKey string
	// NameEncoder collapses the series of Logger names down into one string for logging
	NameEncoder func(names []string) string
	// MessageKey determines the top level JSON object key to store the log message in
	MessageKey string
	// TimestampKey determines the top level JSON object key to store the timestamp in
	TimestampKey string
	// TimestampEncoder formats timestamps into string representations
	TimestampEncoder func(t time.Time) string
	// ErrorKey determines the top level JSON object key to store any error messages in
	ErrorKey string
	// StackTraceKey determines the top level JSON object key to store any stack trace information in
	StackTraceKey string
	// ErrorEncoder  extracts loggable EncodedError information from an error
	ErrorEncoder func(err error) EncodedError
}

// AssertDefaults replaces all uninitialised options with reasonable defaults
func (j *JSONLogSinkOptions) AssertDefaults() {
	if j.Output == nil {
		j.Output = os.Stderr
	}

	if j.SeverityKey == "" {
		j.SeverityKey = DefaultSeverityKey
	}
	if j.SeverityEncoder == nil {
		j.SeverityEncoder = DefaultSeverityEncoder(DefaultSeverity, DefaultErrorSeverity, DefaultSeverityThresholds)
	}

	if j.NameKey == "" {
		j.NameKey = DefaultNameKey
	}
	if j.NameEncoder == nil {
		j.NameEncoder = DefaultNameEncoder(DefaultNameSeparator)
	}

	if j.MessageKey == "" {
		j.MessageKey = DefaultMessageKey
	}

	if j.TimestampKey == "" {
		j.TimestampKey = DefaultTimestampKey
	}
	if j.TimestampEncoder == nil {
		j.TimestampEncoder = DefaultTimestampEncoder(DefaultTimestampFormat)
	}

	if j.ErrorKey == "" {
		j.ErrorKey = DefaultErrorKey
	}
	if j.StackTraceKey == "" {
		j.StackTraceKey = DefaultStackTraceKey
	}
	if j.ErrorEncoder == nil {
		j.ErrorEncoder = DefaultErrorEncoder
	}
}
