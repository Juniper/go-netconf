package netconf

import (
	"testing"
)

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
