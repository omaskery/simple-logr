package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	pkgerrs "github.com/pkg/errors"

	"github.com/omaskery/simple-logr"
)

var (
	ErrItBroke = errors.New("wow it super broke yo")
)

func main() {
	opts := simplelogr.JSONLogSinkOptions{
		Output: os.Stdout,
	}
	opts.AssertDefaults()
	logSink := simplelogr.New(simplelogr.Options{
		Sink:      simplelogr.NewJSONLogSink(opts),
		Verbosity: 10,
	})
	logger := logr.New(logSink).WithName("example").WithValues("hello", "kitty")

	logger.Info("start")
	defer logger.Info("end")

	logger.Info("such a good test", "wow", 10)
	logger.Error(pkgerrs.WithStack(ErrItBroke), "oops", "foo", "flange")
	logger.Error(ErrItBroke, "so upsetting")
	logger.Error(pkgerrs.Wrap(pkgerrs.Wrap(pkgerrs.WithStack(ErrItBroke), "nesting A"), "nesting B"), "so nested")
	logger.Error(fmt.Errorf("go 1.13 wrapping: %w", ErrItBroke), "other wrapping methods")

	logger.V(1).Info("meow", "this", "is a test")
	logger.V(2).Info("woof", "this", "is an even more verbose test")
}
