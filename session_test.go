package netconf

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type testServer struct {
	outMsgs []string
	inMsgs  []string
}

func (ts *testServer) MsgReader() (io.Reader, error) {
	if len(ts.outMsgs) == 0 {
		return nil, io.EOF // XXX: right error?
	}

	r := strings.NewReader(ts.outMsgs[0])
	ts.outMsgs = ts.outMsgs[1:]
	return r, nil
}

func (ts *testServer) MsgWriter() (io.WriteCloser, error) {
	return &testMsgWriter{ts: ts}, nil
}

type testMsgWriter struct {
	ts *testServer
	bytes.Buffer
}

func (w *testMsgWriter) Close() error {
	// XXX: error if previous write not closed
	w.ts.inMsgs = append(w.ts.inMsgs, w.String())
	return nil
}

func (tw *testServer) Close() error { return nil }

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
			ts := &testServer{outMsgs: []string{tc.serverHello}}
			sess := &Session{tr: ts}

			err := sess.hello()
			if err != nil && !tc.shouldError {
				t.Errorf("unexpected error: %v", err)
			}

			if len(ts.inMsgs) != 1 {
				t.Errorf("number of messages written to server wrong (want 1, got %d)", len(ts.inMsgs))
			}

			if sess.SessionID() != tc.wantID {
				t.Errorf("session id does not match (want %q got %q)", tc.wantID, sess.SessionID())
			}
		})
	}
	// tests
}
