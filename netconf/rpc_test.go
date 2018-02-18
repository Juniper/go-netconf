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
