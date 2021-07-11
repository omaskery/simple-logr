# Simple Logr

This little library intended to be a small, customisable component for use with the [logr][logr] Go library.

## Features

* Simple but configurable. Won't be very efficient, but easy to chop and change it to do what you want.
* Integrates well with custom error types like [github.com/pkg/errors][pkgerrs], able to extract stack traces and add
  them to log messages.

## Design

It is broken down into two layers:

* The `Logger` which implements the `logr.LogSink` interface and is responsible for generating log `Entry` objects.
* `LogSink` implementations which are responsible for emitting log `Entry` objects, to wherever they please, formatted
  however they like.
  
There are two provided log sinks:
* `DevelopmentLogSink` - intended for local development convenience, with optionally coloured output
* `JSONLogSink` - structured JSON logging, intended for production

This library hopes to be made of many composable pieces, such that any component that doesn't suit your requirements
can be omitted and replaced. To that end, it uses caller-provided functions where applicable to allow for considerable
flexibility before you are forced to resort writing a new LogSink.

[logr]: https://github.com/go-logr/logr
[pkgerrs]: https://github.com/pkg/errors
