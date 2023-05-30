package netconf

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// RawXML captures the raw xml for the given element.  Used to process certain
// elements later.
type RawXML []byte

func (x *RawXML) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var inner struct {
		Data []byte `xml:",innerxml"`
	}

	if err := d.DecodeElement(&inner, &start); err != nil {
		return err
	}

	*x = inner.Data
	return nil
}

// MarshalXML implements xml.Marshaller.  Raw XML is passed verbatim, errors and
// all.
func (x *RawXML) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	inner := struct {
		Data []byte `xml:",innerxml"`
	}{
		Data: []byte(*x),
	}
	return e.EncodeElement(&inner, start)
}

// helloMsg maps the xml value of the <hello> message in RFC6241
type HelloMsg struct {
	XMLName      xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 hello"`
	SessionID    uint64   `xml:"session-id,omitempty"`
	Capabilities []string `xml:"capabilities>capability"`
}

// RPCMsg maps the xml value of <rpc> in RFC6241
type RPCMsg struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc"`
	MessageID uint64   `xml:"message-id,attr"`
	Operation any      `xml:",innerxml"`
}

func (msg *RPCMsg) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if msg.Operation == nil {
		return fmt.Errorf("operation cannot be nil")
	}

	// TODO: validate operation is named?

	// alias the type to not cause recursion calling e.Encode
	type rpcMsg RPCMsg
	inner := rpcMsg(*msg)
	return e.Encode(&inner)
}

// RPCReplyMsg maps the xml value of <rpc-reply> in RFC6241
type RPCReplyMsg struct {
	XMLName   xml.Name  `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc-reply"`
	MessageID uint64    `xml:"message-id,attr"`
	Errors    RPCErrors `xml:"rpc-error,omitempty"`
	Body      []byte    `xml:",innerxml"`
}

type NotificationMsg struct {
	XMLName   xml.Name  `xml:"urn:ietf:params:xml:ns:netconf:notification:1.0 notification"`
	EventTime time.Time `xml:"eventTime"`
	Data      []byte    `xml:",innerxml"`
}

type ErrSeverity string

const (
	ErrSevError   ErrSeverity = "error"
	ErrSevWarning ErrSeverity = "warning"
)

type ErrType string

const (
	ErrTypeTransport ErrType = "transport"
	ErrTypeRPC       ErrType = "rpc"
	ErrTypeProtocol  ErrType = "protocol"
	ErrTypeApp       ErrType = "app"
)

type ErrTag string

const (
	ErrInUse                 ErrTag = "in-use"
	ErrInvalidValue          ErrTag = "invalid-value"
	ErrTooBig                ErrTag = "too-big"
	ErrMissingAttribute      ErrTag = "missing-attribute"
	ErrBadAttribute          ErrTag = "bad-attribute"
	ErrUnknownAttribute      ErrTag = "unknown-attribute"
	ErrMissingElement        ErrTag = "missing-element"
	ErrBadElement            ErrTag = "bad-element"
	ErrUnknownElement        ErrTag = "unknown-element"
	ErrUnknownNamespace      ErrTag = "unknown-namespace"
	ErrAccesDenied           ErrTag = "access-denied"
	ErrLockDenied            ErrTag = "lock-denied"
	ErrResourceDenied        ErrTag = "resource-denied"
	ErrRollbackFailed        ErrTag = "rollback-failed"
	ErrDataExists            ErrTag = "data-exists"
	ErrDataMissing           ErrTag = "data-missing"
	ErrOperationNotSupported ErrTag = "operation-not-supported"
	ErrOperationFailed       ErrTag = "operation-failed"
	ErrPartialOperation      ErrTag = "partial-operation"
	ErrMalformedMessage      ErrTag = "malformed-message"
)

type RPCError struct {
	Type     ErrType     `xml:"error-type"`
	Tag      ErrTag      `xml:"error-tag"`
	Severity ErrSeverity `xml:"error-severity"`
	AppTag   string      `xml:"error-app-tag,omitempty"`
	Path     string      `xml:"error-path,omitempty"`
	Message  string      `xml:"error-message,omitempty"`
	Info     RawXML      `xml:"error-info,omitempty"`
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
