package transport

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"os"
)

type Prefixer struct {
	prefix string
	writer io.Writer
	nl     bool
	buf    bytes.Buffer // reuse buffer to save allocations
}

// New creates a new Prefixer that forwards all calls to Write() to writer.Write() with all lines prefixed with the
// return value of prefixFunc. Having a function instead of a static prefix allows to print timestamps or other changing
// information.
func newPrefixWriter(writer io.Writer, prefix string) *Prefixer {
	return &Prefixer{prefix: prefix, writer: writer, nl: true}
}

func (pf *Prefixer) Write(payload []byte) (int, error) {
	pf.buf.Reset() // clear the buffer

	for _, b := range payload {
		if pf.nl {
			pf.buf.WriteString(pf.prefix)
			pf.nl = false
		}

		pf.buf.WriteByte(b)

		if b == '\n' {
			// do not print the prefix right after the newline character as this might
			// be the very last character of the stream and we want to avoid a trailing prefix.
			pf.nl = true
		}
	}

	n, err := pf.writer.Write(pf.buf.Bytes())
	if err != nil {
		// never return more than original length to satisfy io.Writer interface
		if n > len(payload) {
			n = len(payload)
		}
		return n, err
	}

	// return original length to satisfy io.Writer interface
	return len(payload), nil
}

type FrameTransport struct {
	r *bufio.Reader
	w *bufio.Writer

	closed bool

	curReader *frameReader
	curWriter *frameWriter
}

func NewFrameTransport(r io.Reader, w io.Writer) *FrameTransport {
	w = io.MultiWriter(w, newPrefixWriter(os.Stdout, "--> "))
	r = io.TeeReader(r, newPrefixWriter(os.Stdout, "<-- "))
	return &FrameTransport{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

func (t *FrameTransport) MsgReader() (io.Reader, error) {
	if t.closed {
		return nil, net.ErrClosed
	}
	// Advance to the start of the next message for the start of the new reader
	if t.curReader != nil {
		if err := t.curReader.advance(); err != nil {
			return nil, err
		}
	}

	t.curReader = &frameReader{t: t}
	return t.curReader, nil
}

var endOfMsg = []byte("]]>]]>")

type frameReader struct {
	t *FrameTransport
}

func (r *frameReader) Read(p []byte) (int, error) {
	// This probably isn't optimal however it looks like xml.Decoder
	// mainly just called ReadByte() and this probably won't ever be
	// used.
	for i := 0; i < len(p); i++ {
		b, err := r.ReadByte()
		if err != nil {
			return i, err
		}
		p[i] = b
	}
	return len(p), nil
}

func (r *frameReader) ReadByte() (byte, error) {
	t := r.t

	if t.closed {
		return 0, io.EOF
	}

	b, err := t.r.ReadByte()
	if err != nil {
		if err == io.EOF && !t.closed {
			return b, io.ErrUnexpectedEOF
		}
		return b, err
	}

	// look for the end of the message marker
	if b == endOfMsg[0] {
		peeked, err := t.r.Peek(len(endOfMsg) - 1)
		if err != nil {
			if err == io.EOF && !t.closed {
				return b, io.ErrUnexpectedEOF
			}
			return b, err
		}

		// check if we are at the end of the message
		if bytes.Equal(peeked, endOfMsg[1:]) {
			t.r.Discard(len(endOfMsg) - 1)
			t.curReader = nil
			return b, io.EOF
		}
	}

	return b, nil
}

// nextFrame will advance to the start of the next framed message.
func (r *frameReader) advance() error {
	for {
		_, err := r.ReadByte()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// MsgWriter allows implementation of a transport.Transport
func (t *FrameTransport) MsgWriter() (io.WriteCloser, error) {
	if t.closed {
		return nil, net.ErrClosed
	}

	if t.curWriter != nil {
		return nil, ErrExistingWriter
	}
	t.curWriter = &frameWriter{t: t}
	return t.curWriter, nil
}

type frameWriter struct {
	t *FrameTransport
}

func (w *frameWriter) Write(p []byte) (int, error) {
	return w.t.w.Write(p)
}

func (w *frameWriter) Close() error {
	t := w.t

	// clear the writer to allow for another one
	t.curWriter = nil

	if err := t.w.WriteByte('\n'); err != nil {
		return err
	}

	if _, err := w.t.w.Write(endOfMsg); err != nil {
		return err
	}

	// Not part of the spec but junos complains when this isn't a newline
	if err := t.w.WriteByte('\n'); err != nil {
		return err
	}

	return t.w.Flush()
}

// XXX: is this needed?
func (t *FrameTransport) Close() error {
	t.closed = true
	return nil
}
