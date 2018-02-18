package netconf

import (
	"testing"
)

func TestRPCErrorError(t *testing.T) {
	rpcErr := RPCError{
		Severity: "lol",
		Message:  "cats",
	}
	expected := "netconf rpc [lol] 'cats'"

	errMsg := rpcErr.Error()
	if errMsg != expected {
		t.Errorf("expected: %s, got: %s", expected, errMsg)
	}
}

func TestMethodLock(t *testing.T) {
	expected := "<lock><target><what.target/></target></lock>"

	mLock := MethodLock("what.target")
	if mLock.MarshalMethod() != expected {
		t.Errorf("got %s, expected %s", mLock, expected)
	}
}

func TestMethodUnlock(t *testing.T) {
	expected := "<unlock><target><what.target/></target></unlock>"

	mUnlock := MethodUnlock("what.target")
	if mUnlock.MarshalMethod() != expected {
		t.Errorf("got %s, expected %s", mUnlock, expected)
	}
}

func TestMethodGetConfig(t *testing.T) {
	expected := "<get-config><source><what.target/></source></get-config>"

	mGetConfig := MethodGetConfig("what.target")
	if mGetConfig.MarshalMethod() != expected {
		t.Errorf("got %s, expected %s", mGetConfig, expected)
	}
}

// TestUUIDLength verifies that UUID length is cor([a-zA-Z]|\d|-)rect
func TestUUIDLength(t *testing.T) {
	expectedLength := 36

	u := uuid()
	actualLength := len(u)
	t.Logf("got UUID: %s", u)
	if actualLength != expectedLength {
		t.Errorf("got wrong length UUID. Expected %d, got %d", expectedLength, actualLength)
	}
}

// TestUUIDChat verifies that UUID contains ASCII letter/number and delimiter
func TestUUIDChar(t *testing.T) {
	//validChars := regexp.MustCompile("([a-zA-Z]|\\d|-)")

	valid := func(i int) bool {
		// A-Z
		if i >= 65 && i <= 90 {
			return true
		}

		// a-z
		if i >= 97 && i <= 122 {
			return true
		}

		// 0-9
		if i >= 48 && i <= 57 {
			return true
		}

		// -
		if i == 45 {
			return true
		}

		return false
	}

	u := uuid()

	for _, v := range u {
		if valid(int(v)) == false {
			t.Errorf("invalid char %s", string(v))

		}
	}
}

var RPCReplytests = []struct {
	rawXML   string
	reply_ok bool
}{
	{
		`
<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" xmlns:junos="http://xml.juniper.net/junos/15.1F4/junos">
<commit-results>
</commit-results>
<ok/>
</rpc-reply>`,
		false,
	},
	{
		`
<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" xmlns:junos="http://xml.juniper.net/junos/15.1F4/junos">
<commit-results>
<rpc-error>
<error-type>application</error-type>
<error-tag>invalid-value</error-tag>
<error-severity>error</error-severity>
<error-path>[edit]</error-path>
<error-message>mgd: Missing mandatory statement: 'root-authentication'</error-message>
<error-info>
<bad-element>system</bad-element>
</error-info>
</rpc-error>
<rpc-error>
<error-type>protocol</error-type>
<error-tag>operation-failed</error-tag>
<error-severity>error</error-severity>
<error-message>
configuration check-out failed: (missing mandatory statements)
</error-message>
</rpc-error>
</commit-results>
</rpc-reply>`,
		false,
	},
	{
		`
<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" xmlns:junos="http://xml.juniper.net/junos/16.1R3/junos">
<commit-results>
<rpc-error>
<error-severity>warning</error-severity>
<error-path>[edit protocols]</error-path>
<error-message>mgd: requires 'mpls' license</error-message>
<error-info>
<bad-element>mpls</bad-element>
</error-info>
</rpc-error>
<rpc-error>
<error-severity>warning</error-severity>
<error-path>[edit protocols]</error-path>
<error-message>mgd: requires 'bgp' license</error-message>
<error-info>
<bad-element>bgp</bad-element>
</error-info>
</rpc-error>
<routing-engine junos:style="normal">
<name>fpc0</name>
<commit-check-success/>
</routing-engine>
</commit-results>
<ok/>
</rpc-reply>`,
		false,
	},
}

func TestNewRPCReply(t *testing.T) {
	for _, tc := range RPCReplytests {
		reply, err := NewRPCReply([]byte(tc.rawXML), false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reply.RawReply != tc.rawXML {
			t.Errorf("NewRPCReply(%q) did not set RawReply to input, got %q", tc.rawXML, reply.RawReply)
		}
	}
}
