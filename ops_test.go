package netconf

import (
	"bytes"
	"context"
	"encoding/xml"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestUnmarshalOk(t *testing.T) {
	tt := []struct {
		name  string
		input string
		want  bool
	}{
		{"selfclosing", "<foo>><ok/></foo>", true},
		{"missing", "<foo></foo>", false},
		{"closetag", "<foo><ok></ok></foo>", true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var v struct {
				XMLName xml.Name   `xml:"foo"`
				Ok      ExtantBool `xml:"ok"`
			}

			if err := xml.Unmarshal([]byte(tc.input), &v); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bool(v.Ok) != tc.want {
				t.Errorf("unexpected results (want: %t got: %t)", tc.want, v.Ok)
			}
		})
	}
}

func TestMarshalDatastore(t *testing.T) {
	tt := []struct {
		input     Datastore
		want      string
		shouldErr bool
	}{
		{Running, "<rpc><target><running/></target></rpc>", false},
		{Startup, "<rpc><target><startup/></target></rpc>", false},
		{Candidate, "<rpc><target><candidate/></target></rpc>", false},
		{Datastore("custom-store"), "<rpc><target><custom-store/></target></rpc>", false},
		{Datastore(""), "", true},
		{Datastore("<xml-elements>"), "<rpc><target><&lt;xml-elements&gt;/></target></rpc>", true},
	}

	for _, tc := range tt {
		t.Run(string(tc.input), func(t *testing.T) {
			v := struct {
				XMLName xml.Name  `xml:"rpc"`
				Target  Datastore `xml:"target"`
			}{Target: tc.input}

			got, err := xml.Marshal(&v)
			if err != nil && !tc.shouldErr {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != tc.want {
				t.Errorf("unexpected results (want:: %q, got: %q)", tc.want, string(got))
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	ts := newTestServer(t)
	sess := newSession(ts.transport())
	go sess.recv()

	ts.queueRespString("<rpc-reply xmlns='urn:ietf:params:xml:ns:netconf:base:1.0' message-id='1'><data>foo</data></rpc-reply>")

	got, err := sess.GetConfig(context.Background(), Running)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = ts.popReqString()
	if err != nil {
		t.Errorf("failed to read request sent to server")
	}

	want := []byte("foo")
	if !bytes.Equal(want, got) {
		t.Errorf("incorrent config returned (want %q, got %q)", want, got)
	}
}

type structuredCfg struct {
	System structuredCfgSystem `xml:"system"`
}

type structuredCfgSystem struct {
	Hostname string `xml:"host-name"`
}

const intfaceConfig = `
<interfaces>
  <interface>
    <name>ge-0/0/2</name>
    <unit>
      <name>0</name>
      <family>
        <inet>
          <address>
            <name>2.2.2.1/32</name>
          </address>
        </inet>
      </family>
    </unit>
  </interface>
</interfaces>
`

func TestEditConfig(t *testing.T) {
	tt := []struct {
		name      string
		target    Datastore
		config    interface{}
		options   []EditConfigOption
		mustMatch []*regexp.Regexp
		noMatch   []*regexp.Regexp
	}{
		{
			name:   "running structured no options",
			target: Running,
			config: structuredCfg{
				System: structuredCfgSystem{
					Hostname: "darkstar",
				},
			},
			mustMatch: []*regexp.Regexp{
				regexp.MustCompile(`<target>\S*<running/>\S*</target>`),
				regexp.MustCompile(
					`<config>\S*<system>\S*<host-name>darkstar</host-name>\S*</system>\S*</config>`,
				),
			},
			noMatch: []*regexp.Regexp{
				regexp.MustCompile(`<url>`),
			},
		},
		{
			name:   "canidate string all options",
			target: Candidate,
			config: intfaceConfig,
			options: []EditConfigOption{
				WithDefaultMergeStrategy(ReplaceConfig),
				WithErrorStrategy(ContinueOnError),
				WithTestStrategy(TestOnly),
			},
			mustMatch: []*regexp.Regexp{
				regexp.MustCompile(`<target>\S*<candidate/>\S*</target>`),
				regexp.MustCompile(`<name>ge-0/0/2</name>`),
				regexp.MustCompile(`<default-operation>replace</default-operation>`),
				regexp.MustCompile(`<test-option>test-only</test-option>`),
				regexp.MustCompile(`<error-option>continue-on-error</error-option>`),
			},
			noMatch: []*regexp.Regexp{
				regexp.MustCompile(`<url>`),
			},
		},
		{
			name:   "byteslice config",
			target: Running,
			config: []byte("<system><services><ssh/></services></system>"),
			mustMatch: []*regexp.Regexp{
				regexp.MustCompile(`<system><services><ssh/></services></system>`),
			},
		},
		{
			name:   "startup url no options",
			target: Startup,
			config: URL("ftp://myftpesrver/foo/config.xml"),
			mustMatch: []*regexp.Regexp{
				regexp.MustCompile(`<target>\S*<startup/>\S*</target>`),
				regexp.MustCompile(`<url>ftp://myftpesrver/foo/config.xml</url>`),
			},
			noMatch: []*regexp.Regexp{
				regexp.MustCompile(`<config>`),
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.EditConfig(context.Background(), tc.target, tc.config, tc.options...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.mustMatch {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match  `%s`", match.String())
				}
			}

			for _, match := range tc.noMatch {
				if match.Match(sentMsg) {
					t.Errorf("sent message matched  `%s`", match.String())
				}
			}

		})
	}
}

// TODO: TestEditConfigError()

func TestCopyConfig(t *testing.T) {
	tt := []struct {
		name           string
		source, target interface{}
		matches        []*regexp.Regexp
	}{
		{
			name:   "running->startup",
			source: Running,
			target: Startup,
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<source>\S*<running/>\S*</source>`),
				regexp.MustCompile(`<target>\S*<startup/>\S*</target>`),
			},
		},
		{
			name:   "running->url",
			source: Running,
			target: URL("ftp://myserver.example.com/router.cfg"),
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<source>\S*<running/>\S*</source>`),
				regexp.MustCompile(`<target>\S*<url>ftp://myserver.example.com/router.cfg</url>\S*</target>`),
			},
		},
		{
			name:   "url->candidate",
			source: URL("http://myserver.example.com/router.cfg"),
			target: Candidate,
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<source>\S*<url>http://myserver.example.com/router.cfg</url>\S*</source>`),
				regexp.MustCompile(`<target>\S*<candidate/>\S*</target>`),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.CopyConfig(context.Background(), tc.source, tc.target)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.matches {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match `%s`", match.String())
				}
			}
		})
	}
}

func TestDeleteConfig(t *testing.T) {
	tt := []struct {
		target  Datastore
		matches []*regexp.Regexp
	}{
		{
			target: Startup,
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<delete-config>\S*<target>\S*<startup/>\S*</target>\S*</delete-config>`),
			},
		},
	}

	for _, tc := range tt {
		t.Run(string(tc.target), func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.DeleteConfig(context.Background(), tc.target)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.matches {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match `%s`", match.String())
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tt := []struct {
		name    string
		source  interface{}
		matches []*regexp.Regexp
	}{
		{
			name:   "candidate",
			source: Candidate,
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<validate>\S*<source>\S*<candidate/>\S*</source>\S*</validate>`),
			},
		},
		// XXX: test []byte,string
		// XXX: test xml object
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.Validate(context.Background(), tc.source)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.matches {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match `%s`", match.String())
				}
			}
		})
	}
}

func TestLock(t *testing.T) {
	tt := []struct {
		target  Datastore
		matches []*regexp.Regexp
	}{
		{
			target: Candidate,
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<lock>\S*<target>\S*<candidate/>\S*</target>\S*</lock>`),
			},
		},
	}

	for _, tc := range tt {
		t.Run(string(tc.target), func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.Lock(context.Background(), tc.target)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.matches {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match `%s`", match.String())
				}
			}
		})
	}
}

func TestUnlock(t *testing.T) {
	tt := []struct {
		target  Datastore
		matches []*regexp.Regexp
	}{
		{
			target: Candidate,
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<unlock>\S*<target>\S*<candidate/>\S*</target>\S*</unlock>`),
			},
		},
	}

	for _, tc := range tt {
		t.Run(string(tc.target), func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.Unlock(context.Background(), tc.target)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.matches {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match `%s`", match.String())
				}
			}
		})
	}
}

func TestKillSession(t *testing.T) {
	tt := []struct {
		id      uint32
		matches []*regexp.Regexp
	}{
		{
			id: 42,
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<kill-session>\S*<session-id>42</session-id>\S*</kill-session>`),
			},
		},
	}

	for _, tc := range tt {
		t.Run(strconv.Itoa(int(tc.id)), func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.KillSession(context.Background(), tc.id)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.matches {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match `%s`", match.String())
				}
			}
		})
	}
}

func TestCommit(t *testing.T) {
	tt := []struct {
		name    string
		options []CommitOption
		matches []*regexp.Regexp
	}{
		{
			name: "noOptions",
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<commit></commit>`),
			},
		},
		{
			name:    "confirmed",
			options: []CommitOption{WithConfirmed()},
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<commit><confirmed></confirmed></commit>`),
			},
		},
		{
			name:    "confirmed",
			options: []CommitOption{WithConfirmedTimeout(1 * time.Minute)},
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<commit><confirmed></confirmed><confirm-timeout>60</confirm-timeout></commit>`),
			},
		},
		{
			name:    "persist",
			options: []CommitOption{WithPersist("myid")},
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<commit><confirmed></confirmed><persist>myid</persist></commit>`),
			},
		},
		{
			name:    "persist_id",
			options: []CommitOption{WithPersistID("myid")},
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<commit><persist-id>myid</persist-id></commit>`),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.Commit(context.Background(), tc.options...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.matches {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match `%s`", match.String())
				}
			}
		})
	}
}

func TestCancelCommit(t *testing.T) {
	tt := []struct {
		name    string
		options []CancelCommitOption
		matches []*regexp.Regexp
	}{
		{
			name: "noOptions",
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<cancel-commit></cancel-commit>`),
			},
		},
		{
			name:    "persist_id",
			options: []CancelCommitOption{WithPersistID("myid")},
			matches: []*regexp.Regexp{
				regexp.MustCompile(`<cancel-commit><persist-id>myid</persist-id></cancel-commit>`),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTestServer(t)
			sess := newSession(ts.transport())
			go sess.recv()

			ts.queueRespString(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><ok/></rpc-reply>`)

			err := sess.CancelCommit(context.Background(), tc.options...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			sentMsg, err := ts.popReq()
			if err != nil {
				t.Errorf("failed to read message sent to sever: %v", err)
			}

			for _, match := range tc.matches {
				if !match.Match(sentMsg) {
					t.Errorf("sent message didn't match `%s`", match.String())
				}
			}
		})
	}
}
