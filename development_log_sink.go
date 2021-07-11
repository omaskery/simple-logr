package simplelogr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/pkg/errors"
)

// DevelopmentLogSink emits unstructured, optionally coloured, text representations of log Entry objects - intended
// for ease of reading in terminals during local development
type DevelopmentLogSink struct {
	options DevelopmentLogSinkOptions
}

// NewDevelopmentLogSink creates a new DevelopmentLogSink with the provided options
func NewDevelopmentLogSink(opts DevelopmentLogSinkOptions) *DevelopmentLogSink {
	sink := &DevelopmentLogSink{
		options: opts,
	}

	allColours := []*color.Color{
		sink.options.PrimaryColour,
		sink.options.SecondaryColour,
	}
	for _, c := range sink.options.SeverityColours {
		allColours = append(allColours, c)
	}

	switch sink.options.ColouredOutput {
	case ColourModeAuto:
		// do nothing, let the color package do its magic
	case ColourModeForceOn:
		for _, c := range allColours {
			c.EnableColor()
		}
	case ColourModeForceOff:
		for _, c := range allColours {
			c.DisableColor()
		}
	}

	return sink
}

// Log implements LogSink, encoding the given Entry as human-readable text before writing it to the configured io.Writer
func (d DevelopmentLogSink) Log(e Entry) error {
	buffer := bytes.Buffer{}

	severity := d.options.SeverityEncoder(e.Level, e.Error)
	severityColour := d.options.SeverityColours[severity]
	if severityColour == nil {
		severityColour = d.options.PrimaryColour
	}

	if _, err := d.options.SecondaryColour.Fprint(&buffer, d.options.TimestampEncoder(e.Timestamp)); err != nil {
		return err
	}

	if _, err := severityColour.Fprintf(&buffer, "%s%s", d.options.SpaceSeparator, severity); err != nil {
		return err
	}

	if len(e.Names) > 0 {
		if _, err := d.options.PrimaryColour.Fprintf(&buffer, "%s%s", d.options.SpaceSeparator, d.options.NameEncoder(e.Names)); err != nil {
			return err
		}
	}

	if _, err := d.options.PrimaryColour.Fprintf(&buffer, "%s%s", d.options.SpaceSeparator, e.Message); err != nil {
		return err
	}

	var encodedErr EncodedError
	if e.Error != nil {
		encodedErr = d.options.ErrorEncoder(e.Error)
		if _, err := severityColour.Fprintf(&buffer, "%s%s=%q", d.options.SpaceSeparator, d.options.ErrorKey, encodedErr.Message); err != nil {
			return err
		}
	}

	for i := 0; i < len(e.KVs); i += 2 {
		k := e.KVs[i]
		v := e.KVs[i+1]

		kStr, ok := k.(string)
		if !ok {
			return errors.Errorf("logging keys must be strings, got %T: %v", k, k)
		}

		if _, err := d.options.SecondaryColour.Fprintf(&buffer, "%s%s=", d.options.SpaceSeparator, kStr); err != nil {
			return err
		}

		b, err := json.Marshal(v)
		if err != nil {
			return err
		}

		if _, err := d.options.PrimaryColour.Fprintf(&buffer, "%s", b); err != nil {
			return err
		}
	}

	if encodedErr.StackTrace != "" {
		if _, err := d.options.PrimaryColour.Fprintf(&buffer, "%s", encodedErr.StackTrace); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(d.options.Output, "%s%s", buffer.String(), d.options.EntrySuffix); err != nil {
		return err
	}

	return nil
}

var _ LogSink = (*DevelopmentLogSink)(nil)

// ColourMode controls whether the DevelopmentLogSink emits coloured output or not
type ColourMode int

const (
	// ColourModeAuto will use the color package's built-in auto-detection to guess whether to use coloured output
	ColourModeAuto ColourMode = iota
	// ColourModeForceOff forces coloured output to be disabled, use this if you're seeing garbled escape characters
	// in the output
	ColourModeForceOff
	// ColourModeForceOn forces coloured output to be produced, use this if you're using an integrated terminal in
	// your development IDE and aren't getting coloured output
	ColourModeForceOn
)

// DevelopmentLogSinkOptions configures the behaviour of a DevelopmentLogSink
type DevelopmentLogSinkOptions struct {
	// Output configures where to write logs to
	Output io.Writer
	// ColouredOutput determines whether coloured output will be used, if unspecified it will attempt to auto-detect
	// from the environment. This is usually confused by integrated terminals in IDEs, so for coloured output in IDEs
	// you may wish to use ColourModeForceOn
	ColouredOutput ColourMode
	// SeverityColours maps severity names (produced by SeverityEncoder) to colours, used when displaying severity names
	// and when Entry objects contain an Entry.Error
	SeverityColours map[string]*color.Color
	// PrimaryColour is the colour of log messages, logger names, and the values of key-value pairs
	PrimaryColour *color.Color
	// SecondaryColour is the colour of timestamps, and the keys of key-value pairs
	SecondaryColour *color.Color
	// SeverityEncoder identifies the severity name based on the verbosity level and the presence of any errors
	SeverityEncoder func(level int, err error) string
	// NameEncoder collapses the series of Logger names down into one string for logging
	NameEncoder func(names []string) string
	// TimestampEncoder formats timestamps into string representations
	TimestampEncoder func(t time.Time) string
	// ErrorKey determines the key prefix on any error messages, displayed as though "just another key-value pair",
	// but (if colours are enabled) printed using the relevant colour (see SeverityColours)
	ErrorKey string
	// ErrorEncoder  extracts loggable EncodedError information from an error
	ErrorEncoder func(err error) EncodedError
	// EntrySuffix is appended to the end of log entries, typically to add a newline between them
	EntrySuffix string
	// SpaceSeparator is placed between all log elements: timestamp, severity, logger name, message, and key-value pairs
	// It can be useful, for example, to change this to "\t" to increase spacing - which may improve readability
	SpaceSeparator string
}

// AssertDefaults replaces all uninitialised options with reasonable defaults
func (d *DevelopmentLogSinkOptions) AssertDefaults() {
	if d.Output == nil {
		d.Output = colorable.NewColorableStdout()
	}

	if d.SeverityColours == nil {
		d.SeverityColours = map[string]*color.Color{}
		for severity, colour := range DefaultSeverityColours {
			colourCopy := *colour
			d.SeverityColours[severity] = &colourCopy
		}
	}

	if d.PrimaryColour == nil {
		colourCopy := *DefaultPrimaryColour
		d.PrimaryColour = &colourCopy
	}

	if d.SecondaryColour == nil {
		colourCopy := *DefaultSecondaryColour
		d.SecondaryColour = &colourCopy
	}

	if d.SeverityEncoder == nil {
		d.SeverityEncoder = DefaultSeverityEncoder(DefaultSeverity, DefaultErrorSeverity, DefaultSeverityThresholds)
	}

	if d.NameEncoder == nil {
		d.NameEncoder = DefaultNameEncoder(DefaultNameSeparator)
	}

	if d.TimestampEncoder == nil {
		d.TimestampEncoder = DefaultTimestampEncoder(DefaultTimestampFormat)
	}

	if d.ErrorKey == "" {
		d.ErrorKey = DefaultErrorKey
	}

	if d.ErrorEncoder == nil {
		d.ErrorEncoder = DefaultErrorEncoder
	}

	if d.EntrySuffix == "" {
		d.EntrySuffix = DefaultEntrySuffix
	}

	if d.SpaceSeparator == "" {
		d.SpaceSeparator = DefaultSpaceSeparator
	}
}
