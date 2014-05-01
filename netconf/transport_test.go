package netconf

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"testing"
)

type nilCloser struct {
	io.Reader
	io.Writer
}

func newNilCloser(r io.Reader, w io.Writer) *nilCloser {
	return &nilCloser{r, w}
}

func (nc *nilCloser) Close() error {
	return nil
}

type transportTest struct {
	transportBasicIO
}

func NewTransportTest(input string) (*transportTest, *bytes.Buffer) {
	testReader := bytes.NewReader([]byte(input))
	testWriter := new(bytes.Buffer)

	var tt transportTest
	tt.ReadWriteCloser = newNilCloser(testReader, testWriter)
	return &tt, testWriter
}

var deviceHello = `<!-- No zombies were killed during the creation of this user interface -->
<!-- user bbennett, class j-super-user -->
<hello>
  <capabilities>
    <capability>urn:ietf:params:xml:ns:netconf:base:1.0</capability>
    <capability>urn:ietf:params:xml:ns:netconf:capability:candidate:1.0</capability>
    <capability>urn:ietf:params:xml:ns:netconf:capability:confirmed-commit:1.0</capability>
    <capability>urn:ietf:params:xml:ns:netconf:capability:validate:1.0</capability>
    <capability>urn:ietf:params:xml:ns:netconf:capability:url:1.0?protocol=http,ftp,file</capability>
    <capability>http://xml.juniper.net/netconf/junos/1.0</capability>
    <capability>http://xml.juniper.net/dmi/system/1.0</capability>
  </capabilities>
  <session-id>19313</session-id>
</hello>
]]>]]>
`

var deviceHelloTests = []struct {
	Name     string
	TestFunc func(*HelloMessage) interface{}
	Expected interface{}
}{
	{
		Name: "SessionID match",
		TestFunc: func(h *HelloMessage) interface{} {
			return h.SessionID
		},
		Expected: 19313,
	},
	{
		Name: "Capability length",
		TestFunc: func(h *HelloMessage) interface{} {
			return len(h.Capabilities)
		},
		Expected: 7,
	},
	{
		Name: "Capability #0",
		TestFunc: func(h *HelloMessage) interface{} {
			return h.Capabilities[0]
		},
		Expected: "urn:ietf:params:xml:ns:netconf:base:1.0",
	},
}

func TestReceiveHello(t *testing.T) {
	tt, _ := NewTransportTest(deviceHello)

	hello, err := tt.ReceiveHello()
	if err != nil {
		t.Errorf("Hello read Error: %s", err)
	}

	for idx, test := range deviceHelloTests {
		result := test.TestFunc(hello)

		if result != test.Expected {
			t.Errorf("#%d: ReceiveHello(%s): Expected: '%#v', Got '%#v'", idx, test.Name, test.Expected, result)
		}
	}
}

var clientHelloTests = []struct {
	Name     string
	TestFunc func(*HelloMessage) interface{}
	Expected interface{}
}{
	{
		Name: "SessionID Nil",
		TestFunc: func(h *HelloMessage) interface{} {
			return h.SessionID
		},
		Expected: 0,
	},
	{
		Name: "Capability length",
		TestFunc: func(h *HelloMessage) interface{} {
			return len(h.Capabilities)
		},
		Expected: 1,
	},
	{
		Name: "Capability #0",
		TestFunc: func(h *HelloMessage) interface{} {
			return h.Capabilities[0]
		},
		Expected: "urn:ietf:params:xml:ns:netconf:base:1.0",
	},
}

func TestSendHello(t *testing.T) {
	tt, out := NewTransportTest("")
	tt.SendHello(&HelloMessage{Capabilities: DEFAULT_CAPABILITIES})
	sentHello := out.String()
	out.Reset()

	hello := new(HelloMessage)
	err := xml.Unmarshal([]byte(sentHello), hello)
	if err != nil {
		t.Errorf("Unmarshal of clientHello XML failed: %s", err)
	}

	for idx, test := range clientHelloTests {
		result := test.TestFunc(hello)

		if result != test.Expected {
			t.Errorf("#%d: SendHello(%s): Expected: '%#v', Got '%#v'", idx, test.Name, test.Expected, result)
		}
	}
}

// Login test needs to be over 4096 bytes to fully test the function
var loginText = `
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

SRX240 (ttyp2)

Password:

--- JUNOS 12.1X45-D15.5 built 2013-09-19 07:42:15 UTC
bbennett@SRX240> 

`

func TestWaitForRegexp(t *testing.T) {
	tt, _ := NewTransportTest(loginText)

	var promptRE = regexp.MustCompile(`([>%])\s+`)
	output, matches, err := tt.WaitForRegexp(promptRE)

	if err != nil {
		t.Errorf("WaitForRegexp failed: %s", err)
	}

	if len(matches) != 1 {
		t.Errorf("WaitForRegexp Length of regexp matches is not equal:  Expecting '%d', got '%d", 1, len(matches))
	}

	if !bytes.Equal(matches[0], []byte(">")) {
		t.Errorf("WaitForRegexp #0 match not equal:  Expecting '%d', got '%d", '>', matches[0])
	}

	if !bytes.Equal(output, []byte(loginText)) {
		t.Errorf("WaitForRegexp output text not equal:  Expecting '%s', got '%s", loginText, output)
	}
}

var waitforStringResponse = `
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

SRX240 (ttyp2)

`

func TestWaitForString(t *testing.T) {
	tt, _ := NewTransportTest(loginText)

	output, err := tt.WaitForString("Password:")

	if err != nil {
		t.Errorf("WaitForString failed: %s", err)
	}

	if output != waitforStringResponse {
		t.Errorf("WaitForRegexp output text not equal:  Expecting '%s', got '%s", waitforStringResponse, output)
	}
}

func TestWaitForBytesEmpty(t *testing.T) {
	tt, _ := NewTransportTest("")

	_, err := tt.WaitForBytes([]byte("Test"))
	if err == nil {
		t.Errorf("WaitForBytes should error on empty input!")
	}
}
