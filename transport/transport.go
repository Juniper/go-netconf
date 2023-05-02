package transport

import (
	"errors"
	"io"
)

var (
	// ErrExistingWriter is returned from MsgWriter when there is already a
	// message io.WriterCloser that hasn't been properly closed yet.
	ErrExistingWriter = errors.New("netconf: existing message writer still open")

	// ErrInvalidIO is returned when a write or read operation is called on
	// message io.Reader or a message io.Writer when they are no longer valid.
	// (i.e a new reader or writer has been obtained)
	ErrInvalidIO = errors.New("netconf: read/write on invalid io")
)

// Transport is used for a netconf.Session to talk to the device.  It is message
// oriented to allow for framing and other details to happen on a per message
// basis.
type Transport interface {
	// MsgReader returns a new io.Reader to read a single netconf message. There
	// can only be a single reader for a transport at a time.  Obtaining a new
	// reader should advance the stream to the start of the next message.`
	MsgReader() (io.ReadCloser, error)

	// MsgWriter returns a new io.WriteCloser to write a single netconf message.
	// After writing a message the writer must be closed. Implementers should
	// make sure only a single writer can be obtained and return a error if
	// multiple writers are attempted.
	MsgWriter() (io.WriteCloser, error)

	// Close will close the underlying transport.
	Close() error
}
