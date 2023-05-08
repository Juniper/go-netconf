package transport

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	rfcChunkedRPC = []byte(`
#4
<rpc
#18
 message-id="102"

#79
     xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <close-session/>
</rpc>
##
`)

	rfcUnchunkedRPC = []byte(`<rpc message-id="102"
     xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <close-session/>
</rpc>`)
)

var chunkedTests = []struct {
	name        string
	input, want []byte
	err         error
}{
	{"normal",
		[]byte("\n#3\nfoo\n##\n"),
		[]byte("foo"),
		nil},
	{"empty frame",
		[]byte("\n##\n"),
		[]byte(""),
		nil},
	{"multichunk",
		[]byte("\n#3\nfoo\n#3\nbar\n##\n"),
		[]byte("foobar"),
		nil},
	{"missing header",
		[]byte("uhoh"),
		[]byte(""),
		ErrMalformedChunk},
	{"eof in header",
		[]byte("\n#\n"),
		[]byte(""),
		io.ErrUnexpectedEOF},
	{"no headler",
		[]byte("\n00\n"),
		[]byte(""),
		ErrMalformedChunk},
	{"malformed header",
		[]byte("\n#big\n"),
		[]byte(""),
		ErrMalformedChunk},
	{"zero len chunk",
		[]byte("\n#0\n"),
		[]byte(""),
		ErrMalformedChunk},
	{"too big chunk",
		[]byte("\n#4294967296\n"),
		[]byte(""),
		ErrMalformedChunk},
	{"rfc example rpc", rfcChunkedRPC, rfcUnchunkedRPC, nil},
}

func TestChunkReaderReadByte(t *testing.T) {
	for _, tc := range chunkedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(bytes.NewReader(tc.input))
			cr := &chunkReader{r: r}

			buf := make([]byte, 8192)

			var (
				b   byte
				n   int
				err error
			)
			for {
				b, err = cr.ReadByte()
				if err != nil {
					break
				}
				buf[n] = b
				n++
			}
			buf = buf[:n]

			if err != io.EOF {
				assert.Equal(t, err, tc.err)
			}
			assert.Equal(t, tc.want, buf)
		})
	}
}

func TestChunkReaderRead(t *testing.T) {
	for _, tc := range chunkedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := &chunkReader{
				r: bufio.NewReader(bytes.NewReader(tc.input)),
			}

			got, err := io.ReadAll(r)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestChunkWriter(t *testing.T) {
	buf := bytes.Buffer{}
	w := &chunkWriter{bufio.NewWriter(&buf)}

	n, err := w.Write([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	n, err = w.Write([]byte("quux"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	err = w.Close()
	assert.NoError(t, err)

	want := []byte("\n#3\nfoo\n#4\nquux\n##\n")
	assert.Equal(t, want, buf.Bytes())
}

func BenchmarkChunkedReadByte(b *testing.B) {
	src := bytes.NewReader(rfcChunkedRPC)
	readers := []struct {
		name string
		r    io.ByteReader
	}{
		// test against bufio as a "baseline"
		{"bufio", bufio.NewReader(src)},
		{"chunkreader", &chunkReader{r: bufio.NewReader(src)}},
	}

	for _, bc := range readers {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = bc.r.ReadByte()
				b.SetBytes(1)
			}
		})
	}
}

func BenchmarkChunkedRead(b *testing.B) {
	src := bytes.NewReader(rfcChunkedRPC)
	readers := []struct {
		name string
		r    io.Reader
	}{
		// test against a standard reader and a bufio for a baseline
		{"bare", onlyReader{src}},
		{"bufio", onlyReader{bufio.NewReader(src)}},
		{"chunkedreader", onlyReader{&chunkReader{r: bufio.NewReader(src)}}},
	}
	dstBuf := &bytes.Buffer{}
	dst := onlyWriter{dstBuf}

	for _, bc := range readers {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				src.Reset(rfcChunkedRPC)
				dstBuf.Reset()
				n, err := io.Copy(&dst, bc.r)
				if err != nil {
					b.Fatal(err)
				}
				b.SetBytes(int64(n))
			}
		})
	}
}

var (
	rfcEOMRPC = []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<rpc message-id="105"
xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <get-config>
    <source><running/></source>
    <config xmlns="http://example.com/schema/1.2/config">
     <users/>
    </config>
  </get-config>
</rpc>
]]>]]>`)
	rfcUnframedRPC = rfcEOMRPC[:len(rfcEOMRPC)-6]
)

var framedTests = []struct {
	name        string
	input, want []byte
	err         error
}{
	{"normal",
		[]byte("foo]]>]]>"),
		[]byte("foo"),
		nil},
	{"empty frame",
		[]byte("]]>]]>"),
		[]byte(""),
		nil},
	{"next message",
		[]byte("foo]]>]]>bar]]>]]>"),
		[]byte("foo"), nil},
	{"no delim",
		[]byte("uhohwhathappened"),
		[]byte("uhohwhathappened"),
		io.ErrUnexpectedEOF},
	{"truncated delim",
		[]byte("foo]]>"),
		[]byte("foo"),
		io.ErrUnexpectedEOF},
	{"partial delim",
		[]byte("foo]]>]]bar]]>]]>"),
		[]byte("foo]]>]]bar"),
		nil},
	{"rfc example rpc", rfcEOMRPC, rfcUnframedRPC, nil},
}

func TestEOMReadByte(t *testing.T) {
	for _, tc := range framedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := &eomReader{
				bufio.NewReader(bytes.NewReader(tc.input)),
			}

			buf := make([]byte, 8192)
			var (
				b   byte
				n   int
				err error
			)
			for {
				b, err = r.ReadByte()
				if err != nil {
					break
				}
				buf[n] = b
				n++
			}
			buf = buf[:n]

			if err != io.EOF {
				assert.Equal(t, err, tc.err)
			}

			assert.Equal(t, tc.want, buf)
		})
	}
}

func TestEOMRead(t *testing.T) {
	for _, tc := range framedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := &eomReader{
				r: bufio.NewReader(bytes.NewReader(tc.input)),
			}
			got, err := io.ReadAll(r)
			assert.Equal(t, err, tc.err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEOMWriter(t *testing.T) {
	buf := bytes.Buffer{}
	w := &eomWriter{w: bufio.NewWriter(&buf)}

	n, err := w.Write([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	err = w.Close()
	assert.NoError(t, err)

	want := []byte("foo\n]]>]]>")
	assert.Equal(t, want, buf.Bytes())
}

// force benchmarks to not use any fancy ReadFroms's or other shortcuts
type onlyReader struct {
	io.Reader
}

// force benchmarks to not use any fancy WriteTo's or other shortcuts
type onlyWriter struct {
	io.Writer
}

func BenchmarkEOMReadByte(b *testing.B) {
	src := bytes.NewReader(rfcEOMRPC)

	readers := []struct {
		name string
		r    io.ByteReader
	}{
		// test against bufio as a "baseline"
		{"bufio", bufio.NewReader(src)},
		{"framereader", &eomReader{r: bufio.NewReader(src)}},
	}

	for _, bc := range readers {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = bc.r.ReadByte()
				b.SetBytes(1)
			}
		})
	}
}

func BenchmarkEOMRead(b *testing.B) {
	src := bytes.NewReader(rfcEOMRPC)

	readers := []struct {
		name string
		r    io.Reader
	}{
		// test against a standard reader and a bufio for a baseline
		{"bare", onlyReader{src}},
		{"bufio", onlyReader{bufio.NewReader(src)}},
		{"framereader", onlyReader{&eomReader{r: bufio.NewReader(src)}}},
	}
	dstBuf := &bytes.Buffer{}
	dst := onlyWriter{dstBuf}

	for _, bc := range readers {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				src.Reset(rfcEOMRPC)
				dstBuf.Reset()
				n, err := io.Copy(&dst, bc.r)
				if err != nil {
					b.Fatal(err)
				}
				b.SetBytes(int64(n))
			}

		})
	}
}
