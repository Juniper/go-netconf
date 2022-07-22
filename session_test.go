package netconf

import (
	"context"
	"testing"
)

func TestRPC(t *testing.T) {
	method := struct {
		XMLName string       `xml:"get-interface-information"`
		Terse   SentinalBool `xml:"terse"`
	}{
		Terse: true,
	}
	s := &Session{}

	if _, err := s.Call(context.Background(), method); err != nil {
		t.Fatal(err)
	}
}
