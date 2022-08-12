package transport

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

type ChunkTransport struct {
	r *bufio.Reader
	w *bufio.Writer
}

func NewChunkTransport(r io.Reader, w io.Writer) *FrameTransport {
	w = io.MultiWriter(w, os.Stdout)
	r = io.TeeReader(r, os.Stdout)
	return &FrameTransport{
		r: bufio.NewReader(r),
		w: bufio.NewWriter(w),
	}
}

var ErrMalformedChunk = errors.New("netconf: invalid chunk")

var endOfChunks = []byte("\n##\n")

// MsgReader allows implementation of a transport.Transport
func (t *ChunkTransport) MsgReader() (io.Reader, error) {
	return &chunkReader{t: t}, nil
}

type chunkReader struct {
	t         *ChunkTransport
	chunkLeft int
}

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

func (r *chunkReader) Read(p []byte) (int, error) {
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

func (r *chunkReader) ReadByte() (byte, error) {
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

// MsgWriter allows implementation of a transport.Transport
func (t *ChunkTransport) MsgWriter() (io.WriteCloser, error) {
	return &chunkWriter{t: t}, nil
}

type chunkWriter struct {
	t *ChunkTransport
}

// Write writes the given bytes to the underlying writer framing it with length as
// defined in RFC6242
func (w *chunkWriter) Write(p []byte) (int, error) {
	if _, err := fmt.Fprintf(w.t.w, "\n#%d\n", len(p)); err != nil {
		return 0, err
	}

	return w.t.w.Write(p)
}

// Close indicated the end of a message and triggers writing the end-of-cunnks string
// to the underlying writer.  This does not and cannot close the underlying transport.
func (w *chunkWriter) Close() error {
	if _, err := w.t.w.Write(endOfChunks); err != nil {
		return err
	}
	return w.t.w.Flush()
}
