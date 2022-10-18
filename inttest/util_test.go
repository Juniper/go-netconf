package inttest

import (
	"bytes"
	"testing"
)

type logWriter struct {
	t      *testing.T
	prefix string
	buf    bytes.Buffer
}

func newLogWriter(prefix string, t *testing.T) *logWriter {
	return &logWriter{
		t:      t,
		prefix: prefix,
	}
}

func (w *logWriter) Write(p []byte) (int, error) {
	for _, ch := range p {
		switch ch {
		case '\r':
			// skip
		case '\n':
			w.t.Log(w.prefix, w.buf.String())
			w.buf.Reset()
		default:
			w.buf.WriteByte(ch)
		}

	}
	return len(p), nil
}

func (w *logWriter) Close() error {
	// flush the rest of the buffer to a log
	if w.buf.Len() > 0 {
		w.t.Log(w.prefix, w.buf.String())
	}
	return nil
}
