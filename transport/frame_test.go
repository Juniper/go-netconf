package transport

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

const rfcFramedRPC = `
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
]]>]]>`

var framedTests = []struct {
	name, input, want string
	err               error
}{
	{"normal", "foo]]>]]>", "foo", nil},
	{"empty frame", "]]>]]>", "", nil},
	{"next message", "foo]]>]]>bar]]>]]>", "foo", nil},
	{"no delim", "uhohwhathappened", "uhohwhathappened", io.ErrUnexpectedEOF},
	{"rfc example rpc", rfcFramedRPC, rfcFramedRPC[:len(rfcFramedRPC)-6], nil},
}

func TestFrameReaderReadByte(t *testing.T) {
	for _, tc := range framedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			fr := NewFrameReader(r)

			buf := make([]byte, 8192)

			var (
				b   byte
				n   int
				err error
			)
			for {
				b, err = fr.ReadByte()
				if err != nil {
					break
				}
				buf[n] = b
				n++
			}
			if err != io.EOF && err != tc.err {
				t.Errorf("unexpected error during read (want: %v, got: %v)", tc.err, err)
			}

			if string(buf[:n]) != tc.want {
				t.Errorf("unexpected read (want: %q, got: %q)", tc.want, buf[:n])
			}
		})
	}
}

func TestFrameReaderRead(t *testing.T) {
	for _, tc := range framedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewFrameReader(bufio.NewReader(strings.NewReader(tc.input)))

			got, err := io.ReadAll(r)
			if err != tc.err {
				t.Errorf("unexpected error during read (want: %v, got: %v)", tc.err, err)
			}

			if string(got) != tc.want {
				t.Errorf("unexpected read (want: %q, got: %q)", tc.want, got)
			}
		})
	}
}

func BenchmarkFrameReaderReadByte(b *testing.B) {
	src := strings.NewReader(rfcFramedRPC)
	readers := []struct {
		name string
		r    io.ByteReader
	}{
		{"bufio", bufio.NewReader(src)},
		{"framereader", NewFrameReader(bufio.NewReader(src))},
	}

	for _, bc := range readers {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bc.r.ReadByte()
				b.SetBytes(1)
			}
		})
	}
}

type onlyReader struct {
	io.Reader
}

type onlyWriter struct {
	io.Writer
}

func BenchmarkRead(b *testing.B) {
	src := strings.NewReader(rfcFramedRPC)
	readers := []struct {
		name string
		r    io.Reader
	}{
		{"bare", onlyReader{src}},
		{"bufio", onlyReader{bufio.NewReader(src)}},
		{"framereader", onlyReader{NewFrameReader(bufio.NewReader(src))}},
	}
	dstBuf := &bytes.Buffer{}
	dst := onlyWriter{dstBuf}

	for _, bc := range readers {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				src.Reset(rfcFramedRPC)
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

const rfcChunkedRPC = `
#4
<rpc
#18
 message-id="102"

#79
     xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <close-session/>
</rpc>
##
`

const rfcUnchunkedRPC = `<rpc message-id="102"
     xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <close-session/>
</rpc>`

var chunkedTests = []struct {
	name, input, want string
	err               error
}{
	{"normal", "\n#3\nfoo\n##\n", "foo", nil},
	{"empty frame", "\n##\n", "", nil},
	{"multichunk", "\n#3\nfoo\n#3\nbar\n##\n", "foobar", nil},
	{"missing header", "uhoh", "", ErrMalformedChunk},
	{"eof in header", "\n#\n", "", io.ErrUnexpectedEOF},
	{"malformed header", "\n00\n", "", ErrMalformedChunk},
	{"malformed header", "\n#big\n", "", ErrMalformedChunk},
	{"rfc example rpc", rfcChunkedRPC, rfcUnchunkedRPC, nil},
}

func TestChunkReaderReadByte(t *testing.T) {
	for _, tc := range chunkedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			cr := NewChunkReader(r)

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
			if err != io.EOF && err != tc.err {
				t.Errorf("unexpected error during read (want: %v, got: %v)", tc.err, err)
			}

			if string(buf[:n]) != tc.want {
				t.Errorf("unexpected read (want: %q, got: %q)", tc.want, buf[:n])
			}
		})
	}
}

func TestChunkReaderRead(t *testing.T) {
	for _, tc := range chunkedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewChunkReader(bufio.NewReader(strings.NewReader(tc.input)))

			got, err := io.ReadAll(r)
			if err != tc.err {
				t.Errorf("unexpected error during read (want: %v, got: %v)", tc.err, err)
			}

			if string(got) != tc.want {
				t.Errorf("unexpected read (want: %q, got: %q)", tc.want, got)
			}
		})
	}
}
