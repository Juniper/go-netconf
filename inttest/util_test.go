package inttest

import (
	"strconv"
	"testing"
)

type logWriter struct {
	t      *testing.T
	prefix string
}

func newLogWriter(prefix string, t *testing.T) *logWriter {
	return &logWriter{
		t:      t,
		prefix: prefix,
	}
}

func (w *logWriter) Write(p []byte) (int, error) {
	w.t.Log(w.prefix, strconv.Quote(string(p)))
	return len(p), nil
}
