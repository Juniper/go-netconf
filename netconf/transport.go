package netconf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
)

const (
	MSG_SEPERATOR = "]]>]]>"
)

var DEFAULT_CAPABILITIES = []string{
	"urn:ietf:params:xml:ns:netconf:base:1.0",
}

type HelloMessage struct {
	XMLName      xml.Name `xml:"hello"`
	Capabilities []string `xml:"capabilities>capability"`
	SessionID    int      `xml:"session-id,omitempty"`
}

type Transport interface {
	Send([]byte) error
	Receive() ([]byte, error)
	Close() error
	ReceiveHello() (*HelloMessage, error)
	SendHello(*HelloMessage) error
}

type transportBasicIO struct {
	io             io.ReadWriteCloser
	chunkedFraming bool
}

func (t *transportBasicIO) Read(b []byte) (int, error) {
	return t.io.Read(b)
}

func (t *transportBasicIO) Write(b []byte) (int, error) {
	return t.io.Write(b)
}

func (t *transportBasicIO) Writeln(b []byte) (int, error) {
	t.io.Write(b)
	t.io.Write([]byte("\n"))
	return 0, nil
}

// Sends a well formated netconf rpc message as a slice of bytes adding on the
// nessisary framining messages.
func (t *transportBasicIO) Send(data []byte) error {
	t.io.Write(data)
	t.io.Write([]byte(MSG_SEPERATOR))
	t.io.Write([]byte("\n"))
	return nil // TODO: Implement error handling!
}

func (t *transportBasicIO) Receive() ([]byte, error) {
	return t.WaitForBytes([]byte(MSG_SEPERATOR))
}

func (t *transportBasicIO) WaitForBytes(m []byte) ([]byte, error) {
	var out bytes.Buffer
	buf := make([]byte, 4096)

	for {
		n, err := t.io.Read(buf)

		if n == 0 {
			return nil, fmt.Errorf("WaitForBytes read no data.")
		}

		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		end := bytes.Index(buf, m)
		if end > -1 {
			out.Write(buf[0:end])
			return out.Bytes(), nil
		}
		out.Write(buf[0:n])
	}

	return nil, fmt.Errorf("WaitForBytes failed")
}

func (t *transportBasicIO) WaitForString(m string) (string, error) {
	out, err := t.WaitForBytes([]byte(m))
	if out != nil {
		return string(out), err
	}
	return "", err
}

func (t *transportBasicIO) WaitForRegexp(re *regexp.Regexp) ([]byte, [][]byte, error) {
	var out bytes.Buffer

	buf := make([]byte, 4096)
	for {
		n, err := t.io.Read(buf)

		if n == 0 {
			break // TODO: Handle Error
		}

		if err != nil {
			if err != io.EOF {
				return nil, nil, err
			}
			break
		}

		loc := re.FindSubmatchIndex(buf)
		if loc != nil {
			out.Write(buf[0:loc[1]])

			var matches [][]byte
			for i := 2; i < len(loc); i += 2 {
				matches = append(matches, buf[loc[i]:loc[i+1]])
			}

			return out.Bytes(), matches, nil
		}
		out.Write(buf[0:n])
	}
	return nil, nil, fmt.Errorf("WaitForRegexp failed")
}

func (t *transportBasicIO) Close() error {
	return t.io.Close()
}

func (t *transportBasicIO) SendHello(hello *HelloMessage) error {
	val, err := xml.MarshalIndent(hello, "  ", "    ")
	if err != nil {
		return err
	}

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

type ReadWriteClose struct {
	reader io.Reader
	writer io.WriteCloser
}

func (r *ReadWriteClose) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

func (r *ReadWriteClose) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func (r *ReadWriteClose) Close() error {
	return r.writer.Close()
}

func NewReadWriteCloser(r io.Reader, w io.WriteCloser) *ReadWriteClose {
	return &ReadWriteClose{reader: r, writer: w}
}
