package simplelogr

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// Logger implements the logr.LogSink interface
type Logger struct {
	info    logr.RuntimeInfo
	options Options
	names   []string
	values  []interface{}
}

// LogSink is a system that accepts log Entry objects and handles them, typically by encoding them and emitting them
// to some destination, e.g. encoding into JSON and then emitting to stdout
type LogSink interface {
	Log(e Entry) error
}

// Options controls the configuration of a new Logger, see New
type Options struct {
	Sink         LogSink
	Verbosity    int
	ErrorHandler func(err error)
}

// New creates a new Logger using the provided Options, applying reasonable defaults where options aren't specified
func New(opts Options) *Logger {
	if opts.Sink == nil {
		sinkOpts := JSONLogSinkOptions{}
		sinkOpts.AssertDefaults()
		opts.Sink = NewJSONLogSink(sinkOpts)
	}

	if opts.ErrorHandler == nil {
		opts.ErrorHandler = DefaultErrorHandler
	}

	return &Logger{
		options: opts,
	}
}

// Init accepts runtime information from the parent logr.Logger
func (l *Logger) Init(info logr.RuntimeInfo) {
	l.info = info
}

// Enabled determines whether this logger would emit Info messages at the specified verbosity level
func (l Logger) Enabled(level int) bool {
	return l.options.Verbosity >= level
}

// Info emits an info level log message
func (l Logger) Info(level int, msg string, keysAndValues ...interface{}) {
	l.log(level, nil, msg, keysAndValues...)
}

// Error emits an log message associated with an error
func (l Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.log(0, err, msg, keysAndValues...)
}

func (l Logger) log(level int, err error, msg string, keysAndValues ...interface{}) {
	now := time.Now().UTC()

	kvsLen := len(l.values) + len(keysAndValues)
	if kvsLen%2 != 0 {
		if err := l.options.Sink.Log(Entry{
			Names:     l.names,
			Timestamp: now,
			Error:     errors.New("odd number of arguments passed as key-value pairs for logging"),
		}); err != nil {
			l.options.ErrorHandler(err)
		}
		return
	}

	kvs := make([]interface{}, kvsLen)
	copy(kvs[:len(l.values)], l.values)
	copy(kvs[len(l.values):], keysAndValues)

	if err := l.options.Sink.Log(Entry{
		Level:     level,
		Names:     l.names,
		Timestamp: now,
		Message:   msg,
		KVs:       kvs,
		Error:     err,
	}); err != nil {
		l.options.ErrorHandler(err)
	}
}

// WithValues produces a new logger containing additional key value pairs
func (l Logger) WithValues(keysAndValues ...interface{}) logr.LogSink {
	l.values = append(l.values, keysAndValues...)
	return &l
}

// WithName produces a new logger with an additional name segment
func (l Logger) WithName(name string) logr.LogSink {
	l.names = append(l.names, name)
	return &l
}

var _ logr.LogSink = (*Logger)(nil)

// Entry represents a log entry prepared by Logger, ready for a LogSink to emit (typically by writing to stdout/stderr)
type Entry struct {
	// Level is the verbosity level of this log event, 0 being "least verbose", and larger numbers being more verbose
	Level int
	// Names is the list of names accumulated by chained calls to Logger.WithName
	Names []string
	// Timestamp is the time the log message was captured
	Timestamp time.Time
	// Message is typically a static string associated with the cause of the log event, often a short explanation of
	// what has occurred
	Message string
	// KVs is a sequence of keys and values, stored [key1, value1, key2, value2, ...], populated by both calls to
	// Logger.WithValues and the keysAndValues arguments to Logger.Info and Logger.Error
	KVs []interface{}
	// Error is the error passed to Logger.Error, and may be nil.
	Error error
}
