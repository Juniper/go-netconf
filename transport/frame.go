package transport

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ErrMalformedChunk represents a message that invalid as defined in the chunk
// framking in RFC6242
var ErrMalformedChunk = errors.New("netconf: invalid chunk")

type frameReader interface {
	io.Reader
	io.ByteReader
	advance() error
}

type frameWriter interface {
	io.WriteCloser
	isClosed() bool
}

// Framer is a wrapper used for transports that implement the framing defined in
// RFC6242.  This supports End-of-Message and Chucked framing methods and
// will move from End-of-Message to Chunked framing after the `Upgrade` method
// has been called.
//
// This is not a transport on it's own (missing the `Close` method) and is
// intended to be embedded into other transports.
type Framer struct {
	r io.Reader
	w io.Writer

	br *bufio.Reader
	bw *bufio.Writer

	curReader frameReader
	curWriter frameWriter

	upgraded bool
}

// NewFramer return a new Framer to be used against the given io.Reader and io.Writer.
func NewFramer(r io.Reader, w io.Writer) *Framer {
	f := &Framer{
		r:  r,
		w:  w,
		br: bufio.NewReader(r),
		bw: bufio.NewWriter(w),
	}

	capDir := os.Getenv("GONETCONF_FRAMED_CAPDIR")
	if capDir != "" {
		if err := os.MkdirAll(capDir, 0o755); err != nil {
			panic(fmt.Sprintf("GO_NETCONF_FRAMER: failed to create capture output dir: %v", err))
		}

		ts := time.Now().Format(time.RFC3339)

		inFilename := filepath.Join(capDir, ts+".in")
		inf, err := os.Create(inFilename)
		if err != nil {
			panic(fmt.Sprintf("failed to create capture file %q: %v", inFilename, err))
		}

		outFilename := filepath.Join(capDir, ts+".out")
		outf, err := os.Create(outFilename)
		if err != nil {
			panic(fmt.Sprintf("failed to create capture file %q: %v", inFilename, err))
		}

		f.DebugCapture(inf, outf)
	}

	return f
}

// DebugCapture will copy all *framed* input/output to the the given
// `io.Writers` for sent or recv data.  Either sent of recv can be nil to not
// capture any data.  Useful for displaying to a screen or capturing to a file
// for debugging.
//
// This needs to be called before `MsgReader` or `MsgWriter`.
func (f *Framer) DebugCapture(in io.Writer, out io.Writer) {
	// XXX: should there be a sentinal flag to indicate write/read has been done already?
	if f.curReader != nil ||
		f.curWriter != nil ||
		f.bw.Buffered() > 0 ||
		f.br.Buffered() > 0 {
		panic("debug capture added with active reader or writer")
	}

	if out != nil {
		f.w = io.MultiWriter(f.w, out)
		f.bw = bufio.NewWriter(f.w)
	}

	if in != nil {
		f.r = io.TeeReader(f.r, in)
		f.br = bufio.NewReader(f.r)
	}
}

// Upgrade will cause the Framer to switch from End-of-Message framing to
// Chunked framing.  This is usually called after netconf exchanged the hello
// messages.
func (t *Framer) Upgrade() {
	// XXX: do we need to protect against race conditions (atomic/mutux?)
	t.upgraded = true
}

// MsgReader returns a new io.Reader that is good for reading exactly one netconf
// message.
//
// Only one reader can be used at a time.  When this is called with an existing
// reader then the underlying reader is avanced to the start of the next message
// and invalidates the old reader before returning a new one.
func (t *Framer) MsgReader() (io.Reader, error) {
	if t.curReader != nil {
		if err := t.curReader.advance(); err != nil {
			return nil, err
		}
	}

	if t.upgraded {
		t.curReader = &chunkReader{r: t.br}
	} else {
		t.curReader = &eomReader{r: t.br}
	}
	return t.curReader, nil
}

// MsgWriter returns an io.WriterCloser that is good for writing exactly one
// netconf message.
//
// One one writer can be used at one time and calling this function with an
// existing, unclosed,  writer will result in an error.
func (t *Framer) MsgWriter() (io.WriteCloser, error) {
	if t.curWriter != nil && !t.curWriter.isClosed() {
		return nil, ErrExistingWriter
	}

	if t.upgraded {
		t.curWriter = &chunkWriter{w: t.bw}
	} else {
		t.curWriter = &eomWriter{w: t.bw}
	}
	return t.curWriter, nil
}

var endOfChunks = []byte("\n##\n")

type chunkReader struct {
	r         *bufio.Reader
	chunkLeft int
}

func (r *chunkReader) advance() error {
	defer func() { r.r = nil }()

	for {
		if r.chunkLeft <= 0 {
			err := r.readHeader()
			switch err {
			case nil:
				break
			case io.EOF:
				return nil
			default:
				return err
			}
		}

		if _, err := r.r.Discard(r.chunkLeft); err != nil {
			return err
		}
	}
}

func (r *chunkReader) readHeader() error {
	peeked, err := r.r.Peek(4)
	switch err {
	case nil:
		break
	case io.EOF:
		return io.ErrUnexpectedEOF
	default:
		return err
	}
	r.r.Discard(2)

	// make sure the preable of `\n#` which is used for both the start of a
	// chuck and the end-of-chunk marker is valid.
	if peeked[0] != '\n' || peeked[1] != '#' {
		return ErrMalformedChunk
	}

	// check to see if we are at the end of the read
	if peeked[2] == '#' && peeked[3] == '\n' {
		r.r.Discard(2)
		return io.EOF
	}

	var n int
	for {
		c, err := r.r.ReadByte()
		if err != nil {
			return err
		}

		if c == '\n' {
			break
		}
		if c < '0' || c > '9' {
			return ErrMalformedChunk
		}
		n = n*10 + int(c) - '0'
	}

	const maxChunk = 4294967295
	if n < 1 || n > maxChunk {
		return ErrMalformedChunk
	}

	r.chunkLeft = n
	return nil
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if r.r == nil {
		return 0, ErrInvalidIO
	}

	// still reading existing chunk
	if r.chunkLeft <= 0 {
		if err := r.readHeader(); err != nil {
			return 0, err
		}
	}

	if len(p) > r.chunkLeft {
		p = p[:r.chunkLeft]
	}

	n, err := r.r.Read(p)
	r.chunkLeft -= n
	return n, err
}

func (r *chunkReader) ReadByte() (byte, error) {
	if r.r == nil {
		return 0, ErrInvalidIO
	}

	// still reading existing chunk
	if r.chunkLeft <= 0 {
		if err := r.readHeader(); err != nil {
			return 0, err
		}
	}

	b, err := r.r.ReadByte()
	if err != nil {
		return 0, err
	}
	r.chunkLeft--
	return b, nil

}

type chunkWriter struct {
	w *bufio.Writer
}

func (w *chunkWriter) Write(p []byte) (int, error) {
	if w.w == nil {
		return 0, ErrInvalidIO
	}

	if _, err := fmt.Fprintf(w.w, "\n#%d\n", len(p)); err != nil {
		return 0, err
	}

	return w.w.Write(p)
}

func (w *chunkWriter) Close() error {
	// poison the writer to prevent writes after close
	defer func() { w.w = nil }()
	if _, err := w.w.Write(endOfChunks); err != nil {
		return err
	}
	return w.w.Flush()
}

func (w *chunkWriter) isClosed() bool { return w.w == nil }

var endOfMsg = []byte("]]>]]>")

type eomReader struct {
	r *bufio.Reader
}

func (r *eomReader) advance() error {
	// poison the reader so that it can no longer be used
	defer func() { r.r = nil }()

	var err error
	for err == nil {
		_, err = r.ReadByte()
		if err == io.EOF {
			return nil
		}
	}
	return err
}

func (r *eomReader) Read(p []byte) (int, error) {
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

func (r *eomReader) ReadByte() (byte, error) {
	if r.r == nil {
		return 0, ErrInvalidIO
	}

	b, err := r.r.ReadByte()
	if err != nil {
		if err == io.EOF {
			return b, io.ErrUnexpectedEOF
		}
		return b, err
	}

	// look for the end of the message marker
	if b == endOfMsg[0] {
		peeked, err := r.r.Peek(len(endOfMsg) - 1)
		if err != nil {
			if err == io.EOF {
				return 0, io.ErrUnexpectedEOF
			}
			return 0, err
		}

		// check if we are at the end of the message
		if bytes.Equal(peeked, endOfMsg[1:]) {
			r.r.Discard(len(endOfMsg) - 1)
			return 0, io.EOF
		}
	}

	return b, nil
}

type eomWriter struct {
	w *bufio.Writer
}

func (w *eomWriter) Write(p []byte) (int, error) {
	if w.w == nil {
		return 0, ErrInvalidIO
	}
	return w.w.Write(p)
}

func (w *eomWriter) Close() error {
	// poison the writer to prevent writes after close
	defer func() { w.w = nil }()

	if err := w.w.WriteByte('\n'); err != nil {
		return err
	}

	if _, err := w.w.Write(endOfMsg); err != nil {
		return err
	}

	return w.w.Flush()
}

func (w *eomWriter) isClosed() bool { return w.w == nil }
