package netconf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
)

const (
	// msgSeperator is used to separate sent messages via NETCONF
	msgSeperator = "]]>]]>"
)

// DefaultCapabilities sets the default capabilities of the client library
var DefaultCapabilities = []string{
	"urn:ietf:params:netconf:base:1.0",
}

// HelloMessage is used when bringing up a NETCONF session
type HelloMessage struct {
	XMLName      xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 hello"`
	Capabilities []string `xml:"capabilities>capability"`
	SessionID    int      `xml:"session-id,omitempty"`
}

// Transport interface defines what characterisitics make up a NETCONF transport
// layer object.
type Transport interface {
	Send([]byte) error
	Receive() ([]byte, error)
	Close() error
	ReceiveHello() (*HelloMessage, error)
	SendHello(*HelloMessage) error
}

type transportBasicIO struct {
	io.ReadWriteCloser
	chunkedFraming bool
}

// Sends a well formated NETCONF rpc message as a slice of bytes adding on the
// nessisary framining messages.
func (t *transportBasicIO) Send(data []byte) error {
	t.Write(data)
	// Pad to make sure the msgSeparator isn't sent across a 4096-byte boundary
	if (len(data)+len(msgSeperator))%4096 < 6 {
		t.Write([]byte("      "))
	}
	t.Write([]byte(msgSeperator))
	t.Write([]byte("\n"))
	return nil // TODO: Implement error handling!
}

func (t *transportBasicIO) Receive() ([]byte, error) {
	return t.WaitForBytes([]byte(msgSeperator))
}

func (t *transportBasicIO) SendHello(hello *HelloMessage) error {
	val, err := xml.Marshal(hello)
	if err != nil {
		return err
	}

	header := []byte(xml.Header)
	val = append(header, val...)
	err = t.Send(val)
	return err
}

func (t *transportBasicIO) ReceiveHello() (*HelloMessage, error) {
	hello := new(HelloMessage)

	val, err := t.Receive()
	if err != nil {
		return hello, err
	}

	err = xml.Unmarshal([]byte(val), hello)
	return hello, err
}

func (t *transportBasicIO) Writeln(b []byte) (int, error) {
	t.Write(b)
	t.Write([]byte("\n"))
	return 0, nil
}

func (t *transportBasicIO) WaitForFunc(f func([]byte) (int, error)) ([]byte, error) {
	var out bytes.Buffer
	buf := make([]byte, 4096)

	pos := 0
	for {
		n, err := t.Read(buf[pos : pos+(len(buf)/2)])
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		if n > 0 {
			end, err := f(buf[0 : pos+n])
			if err != nil {
				return nil, err
			}

			if end > -1 {
				out.Write(buf[0:end])
				return out.Bytes(), nil
			}

			if pos > 0 {
				out.Write(buf[0:pos])
				copy(buf, buf[pos:pos+n])
			}

			pos = n
		}
	}

	return nil, fmt.Errorf("WaitForFunc failed")
}

func (t *transportBasicIO) WaitForBytes(b []byte) ([]byte, error) {
	return t.WaitForFunc(func(buf []byte) (int, error) {
		return bytes.Index(buf, b), nil
	})
}

func (t *transportBasicIO) WaitForString(s string) (string, error) {
	out, err := t.WaitForBytes([]byte(s))
	if out != nil {
		return string(out), err
	}
	return "", err
}

func (t *transportBasicIO) WaitForRegexp(re *regexp.Regexp) ([]byte, [][]byte, error) {
	var matches [][]byte
	out, err := t.WaitForFunc(func(buf []byte) (int, error) {
		loc := re.FindSubmatchIndex(buf)
		if loc != nil {
			for i := 2; i < len(loc); i += 2 {
				matches = append(matches, buf[loc[i]:loc[i+1]])
			}
			return loc[1], nil
		}
		return -1, nil
	})
	return out, matches, err
}

// ReadWriteCloser represents a combined IO Reader and WriteCloser
type ReadWriteCloser struct {
	io.Reader
	io.WriteCloser
}

// NewReadWriteCloser creates a new combined IO Reader and Write Closer from the
// provided objects
func NewReadWriteCloser(r io.Reader, w io.WriteCloser) *ReadWriteCloser {
	return &ReadWriteCloser{r, w}
}
