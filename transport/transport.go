package transport

import (
	"errors"
	"io"
)

var ErrExistingWriter = errors.New("netconf: existing message writer still open")

type Transport interface {
	// MsgReader returns a new io.Reader to read a single netconf message.
	// There can only be a single reader for a transport at a time.  Obtaining
	// a new reader should advance the stream to the start of the next message.`
	MsgReader() (io.Reader, error)

	// MsgWriter returns a new io.WriteCloser to write a single netconf
	// message.  After writing a message the writer must be closed.
	// Implementers should make sure only a single writer can be obtained and
	// return a error if multiple writers are attempted.
	MsgWriter() (io.WriteCloser, error)

	// Close will close the transport.  Any existing readers or writers should
	// return (XXX: what error?) after the transport is closed.
	Close() error
}

type Upgrader interface {
	// XXX: this should take a version and/or return an error?
	Upgrade()
}
