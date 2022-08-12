package transport

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

var (
	rfcFramedRPC = []byte(`
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
	rfcUnframedRPC = rfcFramedRPC[:len(rfcFramedRPC)-6]
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
	{"rfc example rpc", rfcFramedRPC, rfcUnframedRPC, nil},
}

func TestFrameReaderReadByte(t *testing.T) {
	for _, tc := range framedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(bytes.NewReader(tc.input))
			mr, err := NewFrameTransport(r, nil).MsgReader()
			if err != nil {
				t.Errorf("failed to get message reader: %v", err)
			}

			fr := mr.(*frameReader)

			buf := make([]byte, 8192)

			var (
				b byte
				n int
			)
			for {
				b, err = fr.ReadByte()
				if err != nil {
					break
				}
				buf[n] = b
				n++
			}
			buf = buf[:n]

			if err != io.EOF && err != tc.err {
				t.Errorf("unexpected error during read (want: %v, got: %v)", tc.err, err)
			}

			if !bytes.Equal(buf, tc.want) {
				t.Errorf("unexpected read (want: %q, got: %q)", tc.want, buf)
			}
		})
	}
}

func TestFrameReaderRead(t *testing.T) {
	for _, tc := range framedTests {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(bytes.NewReader(tc.input))
			mr, err := NewFrameTransport(r, nil).MsgReader()
			if err != nil {
				t.Errorf("failed to get message reader: %v", err)
			}

			got, err := io.ReadAll(mr)
			if err != tc.err {
				t.Errorf("unexpected error during read (want: %v, got: %v)", tc.err, err)
			}

			if !bytes.Equal(got, tc.want) {
				t.Errorf("unexpected read (want: %q, got: %q)", tc.want, got)
			}
		})
	}
}

func TestFrameWriter(t *testing.T) {
	buf := bytes.Buffer{}
	mw, err := NewFrameTransport(nil, bufio.NewWriter(&buf)).MsgWriter()
	if err != nil {
		t.Fatalf("failed to get message writer: %v", err)
	}

	n, err := mw.Write([]byte("foo"))
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	if n != 3 {
		t.Errorf("failed number of bytes written (got %d, want %d)", n, 3)
	}

	if err := mw.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	want := []byte("foo]]>]]>\n")
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("unexpected data written (want %q, got %q", want, buf.Bytes())
	}
}

// force benchmarks to not use any fancy ReadFroms's or other shortcuts
type onlyReader struct {
	io.Reader
}

// force benchmarks to not use any fancy WriteTo's or other shortcuts
type onlyWriter struct {
	io.Writer
}

func BenchmarkFrameReaderReadByte(b *testing.B) {
	src := bytes.NewReader(rfcFramedRPC)
	mr, err := NewFrameTransport(bufio.NewReader(src), nil).MsgReader()
	if err != nil {
		b.Fatalf("failed to get msg reader: %v", err)
	}
	fr := mr.(*frameReader)

	readers := []struct {
		name string
		r    io.ByteReader
	}{
		// test against bufio as a "baseline"
		{"bufio", bufio.NewReader(src)},
		{"framereader", fr},
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

func BenchmarkFrameReaderRead(b *testing.B) {
	src := bytes.NewReader(rfcFramedRPC)
	mr, err := NewFrameTransport(bufio.NewReader(src), nil).MsgReader()
	if err != nil {
		b.Fatalf("failed to get msg reader: %v", err)
	}

	readers := []struct {
		name string
		r    io.Reader
	}{
		// test against a standard reader and a bufio for a baseline
		{"bare", onlyReader{src}},
		{"bufio", onlyReader{bufio.NewReader(src)}},
		{"framereader", onlyReader{mr}},
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
