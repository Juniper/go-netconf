package netconf

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// helloMsg maps the xml value of the <hello> message in RFC6241
type HelloMsg struct {
	XMLName      xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 hello"`
	SessionID    uint64   `xml:"session-id,omitempty"`
	Capabilities []string `xml:"capabilities>capability"`
}

// RPCMsg maps the xml value of <rpc> in RFC6241
type RPCMsg struct {
	XMLName   xml.Name    `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc"`
	MessageID uint64      `xml:"message-id,attr"`
	Operation interface{} `xml:",innerxml"`
}

// RPCReplyMsg maps the xml value of <rpc-reply> in RFC6241
type RPCReplyMsg struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc-reply"`
	MessageID uint64   `xml:"message-id,attr"`

	// Ok is part of RFC6241 and is present if no data is returned from an
	// RPC call and there were no errors.  This IS NOT set to true if data is
	// also returned.  To check if a call is ok then look at the Errors field

	Errors RPCErrors `xml:"rpc-error,omitempty"`
	Data   []byte    `xml:",innerxml"`
}

type NotificationMsg struct {
	XMLName   xml.Name  `xml:"urn:ietf:params:xml:ns:netconf:notification:1.0 notification"`
	EventTime time.Time `xml:"eventTime"`
	Data      []byte    `xml:",innerxml"`
}

type ErrSeverity string

const (
	SevError   ErrSeverity = "error"
	SevWarning ErrSeverity = "warning"
)

type ErrType string

const (
	ErrTypeTrans ErrType = "transport"
	ErrTypeRPC   ErrType = "rpc"
	ErrTypeProto ErrType = "protocol"
	ErrTypeApp   ErrType = "app"
)

const ErrTypeTransport ErrType = "transport"

type RPCError struct {
	Type     string      `xml:"error-type"`
	Tag      string      `xml:"error-tag"`
	Severity ErrSeverity `xml:"error-severity"`
	AppTag   string      `xml:"error-app-tag,omitempty"`
	Path     string      `xml:"error-path,omitempty"`
	Message  string      `xml:"error-message,omitempty"`
	Info     interface{} `xml:"error-info,omitempty"`
}

func (e RPCError) Error() string {
	return fmt.Sprintf("rpc error: %s", e.Message)
}

type RPCErrors []RPCError

func (errs RPCErrors) Error() string {
	var sb strings.Builder
	for i, err := range errs {
		if i > 0 {
			sb.WriteRune('\n')
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

func (errs RPCErrors) Unwrap() []error {
	boxed := make([]error, len(errs))
	for i, err := range errs {
		boxed[i] = err
	}
	return boxed
}
