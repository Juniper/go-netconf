package transport

import (
	"io"
)

type Transport interface {
	MsgReader() io.Reader
	MsgWriter() io.WriteCloser
	Close() error
}
