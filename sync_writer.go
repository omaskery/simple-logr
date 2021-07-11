package simplelogr

import (
	"io"
	"sync"
)

// SynchronizedWriter is used for wrapping an io.Writer with a sync.Mutex so that it is thread safe, e.g. when
// wrapping file handles.
type SynchronizedWriter struct {
	Underlying io.Writer
	lock       sync.Mutex
}

// SynchronizeWritesTo wraps an io.Writer, producing a new io.Writer that is thread-safe
func SynchronizeWritesTo(w io.Writer) *SynchronizedWriter {
	return &SynchronizedWriter{
		Underlying: w,
	}
}

// Write implements io.Writer, acquiring a mutex to prevent concurrency issues when writing to the underlying io.Writer
func (s *SynchronizedWriter) Write(p []byte) (n int, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Underlying.Write(p)
}

var _ io.Writer = (*SynchronizedWriter)(nil)
