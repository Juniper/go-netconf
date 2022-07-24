package transport

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

type FramedTransport struct {
	r *bufio.Reader
	w *bufio.Writer

	upgraded bool
}

func NewFramedTransport(r io.Reader, w io.WriteCloser) *FramedTransport {
	return &FramedTransport{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

// MsgWriter allows implementation of a transport.Transport
func (t *FramedTransport) MsgWriter() io.WriteCloser {
	if t.upgraded {
		return NewChunkWriter(t.w)
	}
	return NewFrameWriter(t.w)
}

// MsgReader allows implementation of a transport.Transport
func (t *FramedTransport) MsgReader() io.Reader {
	if t.upgraded {
		return NewChunkReader(t.r)
	}
	return NewFrameReader(t.r)
}

var endOfMsg = []byte("]]>]]>")

type FrameReader struct {
	r *bufio.Reader
}

func NewFrameReader(r *bufio.Reader) *FrameReader {
	return &FrameReader{r: r}
}

func (r *FrameReader) Read(p []byte) (int, error) {
	// This probably isn't optimal however it looks like xml.Decoder
	// mainly just called ReadByte() and this probably won't even be
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

func (r *FrameReader) ReadByte() (byte, error) {
	peeked, err := r.r.Peek(len(endOfMsg))
	if err != nil && err != io.EOF {
		// if we get an io.EOF here it means we will never read the endOfMsg
		// terminator and will eventually get an EOF but it's better to get
		// there naturually than to return ErrUnexpectedEOF early.
		return 0, err
	}

	// check if we are at the end of the message
	if bytes.Equal(peeked, endOfMsg) {
		r.r.Discard(len(endOfMsg))
		return 0, io.EOF
	}

	b, err := r.r.ReadByte()
	// if we got an EOF before reading the end of marker it should be an error.
	if err == io.EOF {
		return b, io.ErrUnexpectedEOF
	}
	return b, err
}

type FrameWriter struct {
	w *bufio.Writer
}

func NewFrameWriter(w *bufio.Writer) *FrameWriter {
	return &FrameWriter{w: w}
}

func (w *FrameWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *FrameWriter) Close() error {
	if _, err := w.w.Write(endOfMsg); err != nil {
		return err
	}
	// Not part of the spec but junos complains when this isn't a newline
	w.w.WriteByte('\n')
	return w.w.Flush()
}

var ErrMalformedChunk = errors.New("netconf: invalid chunk")

var endOfChunks = []byte("\n##\n")

type ChunkReader struct {
	r         *bufio.Reader
	chunkLeft int
}

func NewChunkReader(r *bufio.Reader) *ChunkReader {
	return &ChunkReader{r: r}
}

func (r *ChunkReader) readHeader() error {

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

func (r *ChunkReader) Read(p []byte) (int, error) {
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

func (r *ChunkReader) ReadByte() (byte, error) {
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

type ChunkWriter struct {
	w *bufio.Writer
}

func NewChunkWriter(w *bufio.Writer) *ChunkWriter {
	return &ChunkWriter{w: w}
}

// Write writes the given bytes to the underlying writer framing it with length as
// defined in RFC6242
func (w *ChunkWriter) Write(p []byte) (int, error) {
	if _, err := fmt.Fprintf(w.w, "\n#%d\n", len(p)); err != nil {
		return 0, err
	}

	return w.w.Write(p)
}

// Close indicated the end of a message and triggers writing the end-of-cunnks string
// to the underlying writer.  This does not and cannot close the underlying transport.
func (w *ChunkWriter) Close() error {
	if _, err := w.w.Write(endOfChunks); err != nil {
		return err
	}
	return w.w.Flush()
}
