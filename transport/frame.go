package transport

import (
	"bufio"
	"bytes"
	"io"
)

// FrameTransport wraps an underlying reader/writer to support the legacy framed
// netconf transport in which each message is deliminated by the end-of-message
// marker of `[[>[[>` as defined in RFC4742 (and again in RFC6242 for support
// for older netconf implementations)
type FrameTransport struct {
	r *bufio.Reader
	w *bufio.Writer

	curReader *frameReader
	curWriter *frameWriter
}

// NewFrameTransport returns a FrameTransport for the given reader and writer.
func NewFrameTransport(r io.Reader, w io.Writer) *FrameTransport {
	return &FrameTransport{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

func (t *FrameTransport) Close() error {
	// XXX: should we error on getting a new msgreader/msgwriter when this is closed or
	// is relying on the underlying transport good enough?
	return nil
}

// MsgReader returns a new io.Reader that is good for reading exactly one netconf
// message that has been framed useing the End-of-Message marker.
//
// Only one reader can be used at a time.  When this is called with an existing
// reader then the underlying reader is avanced to the start of the next message
// and invalidates the old reader beffore returning a new one.
func (t *FrameTransport) MsgReader() (io.Reader, error) {
	// Advance to the start of the next message for the start of the new reader
	if t.curReader != nil {
		if err := t.curReader.advance(); err != nil {
			return nil, err
		}
		t.curReader.expired = true
		t.curReader = nil
	}

	t.curReader = &frameReader{t: t}
	return t.curReader, nil
}

var endOfMsg = []byte("]]>]]>")

type frameReader struct {
	t       *FrameTransport
	expired bool
}

// advance will consume the rest of the current message to prepare for a new
// reader starting with a new netconf message.
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

// Read implements an io.Reader and will read the data inside the current
// message.
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

// ReadByte will read a single byte from the current message.  This is used by a
// xml.Decoder object as not to double buffer the incoming stream.
func (r *frameReader) ReadByte() (byte, error) {
	if r.expired {
		return 0, ErrReaderExpired
	}

	t := r.t
	b, err := t.r.ReadByte()
	if err != nil {
		if err == io.EOF {
			return b, io.ErrUnexpectedEOF
		}
		return b, err
	}

	// look for the end of the message marker
	if b == endOfMsg[0] {
		peeked, err := t.r.Peek(len(endOfMsg) - 1)
		if err != nil {
			if err == io.EOF {
				return b, io.ErrUnexpectedEOF
			}
			return b, err
		}

		// check if we are at the end of the message
		if bytes.Equal(peeked, endOfMsg[1:]) {
			t.r.Discard(len(endOfMsg) - 1)
			return b, io.EOF
		}
	}

	return b, nil
}

// MsgWriter returns an io.WriterCloser that is good for writing exactly one
// netconf message that is framed with the End-of-message marker.
//
// One one writer can be used at one time and calling this function with an
// existing, unclosed,  writer will result in an error.
func (t *FrameTransport) MsgWriter() (io.WriteCloser, error) {
	if t.curWriter != nil {
		return nil, ErrExistingWriter
	}

	t.curWriter = &frameWriter{t: t}
	return t.curWriter, nil
}

type frameWriter struct {
	t *FrameTransport
}

// Write writes the given bytes to the underlying writerv verbatim.
func (w *frameWriter) Write(p []byte) (int, error) {
	return w.t.w.Write(p)
}

// Close indicated the end of a netconf message and triggers writing the
// end-of-mssages terminator (`\n]]>]]>\n`).  This does not (and cannot) close
// the underlying writer.
func (w *frameWriter) Close() error {
	t := w.t

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

	// clear the writer to allow for another one
	t.curWriter = nil

	return t.w.Flush()
}
