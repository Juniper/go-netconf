package transport

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// ChunkTransport wraps an underling reader/writer to support chunked framing
// defined in RFC6242.
type ChunkTransport struct {
	r *bufio.Reader
	w *bufio.Writer

	curReader *chunkReader
	curWriter *chunkWriter
}

// NewChunkTransport returns a new ChunkTransport for the given reader and
// writer.
func NewChunkTransport(r io.Reader, w io.Writer) *ChunkTransport {
	/*
		w = io.MultiWriter(w, os.Stdout)
		r = io.TeeReader(r, os.Stdout)
	*/
	return &ChunkTransport{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

func (t *ChunkTransport) Close() error {
	// XXX: should we error on getting a new msgreader/msgwriter when this is closed or
	// is relying on the underlying transport good enough?

	return nil
}

// ErrMalformedChunk represents a message that invalid as defined in the chunk
// framking in RFC6242
var ErrMalformedChunk = errors.New("netconf: invalid chunk")

var endOfChunks = []byte("\n##\n")

// MsgReader returns a new io.Reader that is good for reading exactly one netconf
// message that has been framed using chunks.
//
// Only one reader can be used at a time.  When this is called with an existing
// reader then the underlying reader is avanced to the start of the next message
// and invalidates the old reader beffore returning a new one.
func (t *ChunkTransport) MsgReader() (io.Reader, error) {
	if t.curReader != nil {
		if err := t.curReader.advance(); err != nil {
			return nil, err
		}
		t.curReader.expired = true
		t.curReader = nil
	}

	t.curReader = &chunkReader{t: t}
	return t.curReader, nil
}

type chunkReader struct {
	t         *ChunkTransport
	chunkLeft int
	expired   bool
}

// advance will consume the rest of the current message to prepare for a new
// reader starting with a new netconf message.
func (r *chunkReader) advance() error {
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

		if _, err := r.t.r.Discard(r.chunkLeft); err != nil {
			return err
		}
	}
}

// readHeader will read the chunk header which contains the length of the
// existing chunk.  It will return EOF if we get the End-of-Chunk header
// (`\n##\n`)
func (r *chunkReader) readHeader() error {
	t := r.t
	peeked, err := t.r.Peek(4)
	switch err {
	case nil:
		break
	case io.EOF:
		return io.ErrUnexpectedEOF
	default:
		return err
	}
	t.r.Discard(2)

	// make sure the preable of `\n#` which is used for both the start of a
	// chuck and the end-of-chunk marker is valid.
	if peeked[0] != '\n' || peeked[1] != '#' {
		return ErrMalformedChunk
	}

	// check to see if we are at the end of the read
	if peeked[2] == '#' && peeked[3] == '\n' {
		t.r.Discard(2)
		return io.EOF
	}

	var n int
	for {
		c, err := t.r.ReadByte()
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

// Read implements an io.Reader and will read the data inside the chunk.
func (r *chunkReader) Read(p []byte) (int, error) {
	if r.expired {
		return 0, ErrReaderExpired
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

	n, err := r.t.r.Read(p)
	r.chunkLeft -= n
	return n, err
}

// ReadByte will read a single byte from the current message.  This is used by a
// xml.Decoder object as not to double buffer the incoming stream.
func (r *chunkReader) ReadByte() (byte, error) {
	if r.expired {
		return 0, ErrReaderExpired
	}

	// still reading existing chunk
	if r.chunkLeft <= 0 {
		if err := r.readHeader(); err != nil {
			return 0, err
		}
	}

	b, err := r.t.r.ReadByte()
	if err != nil {
		return 0, err
	}
	r.chunkLeft--
	return b, nil

}

// MsgWriter returns an io.WriterCloser that is good for writing exactly one
// netconf message that is framed using chunk framing.
//
// One one writer can be used at one time and calling this function with an
// existing, unclosed,  writer will result in an error.
func (t *ChunkTransport) MsgWriter() (io.WriteCloser, error) {
	if t.curWriter != nil {
		return nil, ErrExistingWriter
	}

	return &chunkWriter{t: t}, nil
}

type chunkWriter struct {
	t *ChunkTransport
}

// Write writes the given bytes to the underlying writer framking each method
// as a chunk.
func (w *chunkWriter) Write(p []byte) (int, error) {
	if _, err := fmt.Fprintf(w.t.w, "\n#%d\n", len(p)); err != nil {
		return 0, err
	}

	return w.t.w.Write(p)
}

// Close indicated the end of a netconf message and triggers writing the
// end-of-cunnks terminatlor (`\n##\n`) This does not (and cannot) close the
// underlying transport.
func (w *chunkWriter) Close() error {
	if _, err := w.t.w.Write(endOfChunks); err != nil {
		return err
	}
	return w.t.w.Flush()
}
