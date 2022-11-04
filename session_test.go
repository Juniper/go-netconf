package netconf

import (
	"fmt"
	"io"
	"testing"
)

type testServer struct {
	t   *testing.T
	in  chan []byte
	out chan []byte
}

func newTestServer(t *testing.T) *testServer {
	return &testServer{
		t:   t,
		in:  make(chan []byte),
		out: make(chan []byte),
	}
}

func (s *testServer) handle(r io.Reader, w io.WriteCloser) {
	in, err := io.ReadAll(r)
	if err != nil {
		panic(fmt.Sprintf("testerver: failed to read incomming message: %v", err))
	}
	s.t.Logf("testserver recv: %s", in)
	go func() { s.in <- in }()

	out, ok := <-s.out
	if !ok {
		panic("testserver: no message to send")
	}
	s.t.Logf("tesserver send: %s", out)

	_, err = w.Write(out)
	if err != nil {
		panic(fmt.Sprintf("testserver: failed to write message: %v", err))
	}

	if err := w.Close(); err != nil {
		panic("tesserver: failed to close outbound message")
	}
}

func (s *testServer) queueResp(p []byte)         { go func() { s.out <- p }() }
func (s *testServer) queueRespString(str string) { s.queueResp([]byte(str)) }
func (s *testServer) popReq() ([]byte, error) {
	msg, ok := <-s.in
	if !ok {
		return nil, fmt.Errorf("testserver: no message to read:")
	}
	return msg, nil
}

func (s *testServer) popReqString() (string, error) {
	p, err := s.popReq()
	return string(p), err
}

func (s *testServer) transport() *testTransport { return newTestTransport(s.handle) }

type testTransport struct {
	handler func(r io.Reader, w io.WriteCloser)
	out     chan io.Reader
	// msgReceived, msgSent int
}

func newTestTransport(handler func(r io.Reader, w io.WriteCloser)) *testTransport {
	return &testTransport{
		handler: handler,
		out:     make(chan io.Reader),
	}
}

func (s *testTransport) MsgReader() (io.Reader, error) {
	return <-s.out, nil
}

func (s *testTransport) MsgWriter() (io.WriteCloser, error) {
	inr, inw := io.Pipe()
	outr, outw := io.Pipe()

	go func() { s.out <- outr }()
	go s.handler(inr, outw)

	return inw, nil
}

func (s *testTransport) Close() error {
	if len(s.out) > 0 {
		return fmt.Errorf("testtransport: remaining outboard messages not sent at close")
	}
	return nil
}

const (
	helloGood = `
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <capabilities>
	<capability>urn:ietf:params:netconf:base:1.0</capability>
	<capability>urn:ietf:params:netconf:base:1.1</capability>
  </capabilities>
  <session-id>42</session-id>
</hello>`

	helloBadXML = `
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <capabilities//>
</hello>`

	helloNoSessID = `
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <capabilities>
	<capability>urn:ietf:params:netconf:base:1.0</capability>
	<capability>urn:ietf:params:netconf:base:1.1</capability>
  </capabilities>
</hello>`

	helloNoCaps = `
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <capabilities></capabilities>
  <session-id>42</session-id>
</hello>`
)

func TestHello(t *testing.T) {
	tt := []struct {
		name        string
		serverHello string
		shouldError bool
		wantID      uint64
	}{
		{"good", helloGood, false, 42},
		{"bad xml", helloBadXML, true, 0},
		{"no capabilities", helloNoCaps, true, 0},
		{"no session-id", helloNoSessID, true, 0},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer(t)
			sess := &Session{tr: ts.transport()}

			ts.queueRespString(tc.serverHello)

			err := sess.hello()
			if err != nil && !tc.shouldError {
				t.Errorf("unexpected error: %v", err)
			}

			_, err = ts.popReqString()
			if err != nil {
				t.Error("failed to get response")
			}

			if sess.SessionID() != tc.wantID {
				t.Errorf("session id does not match (want: %q got: %q)", tc.wantID, sess.SessionID())
			}
		})
	}
}
