package netconf

import (
	"encoding/xml"
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
