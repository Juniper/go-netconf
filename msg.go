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
	SessionID    int      `xml:"session-id,omitempty"`
	Capabilities []string `xml:"capabilities>capability"`
}

// rpcMsg maps the xml value of <rpc> in RFC6241
type RPCMsg struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc"`
	MessageID uint64   `xml:"message-id,attr"`
	Operation any      `xml:",innerxml"`
}

// rpcReplyMsg maps the xml value of <rpc-reply> in RFC6241
type RPCReplyMsg struct {
	XMLName   xml.Name   `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc-reply"`
	MessageID uint64     `xml:"message-id,attr"`
	Ok        bool       `xml:"ok"`
	Errors    []RPCError `xml:"rpc-error,omitempty"`
	Data      []byte     `xml:",innerxml"`
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

type Capability string

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
	Info     any         `xml:"error-info,omitempty"`
}

func (e RPCError) Error() string {
	return e.Message
}

type Filter struct {
	XMLName xml.Name `xml:"filter"`
	Type    xml.Attr `xml:"type"`
}

type GetConfigOp struct {
	XMLName xml.Name   `xml:"get-config"`
	Source  StringElem `xml:"source"`
	Filter  *Filter    `xml:"filter,omitempty"`
}

type StringElem string

func (s StringElem) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if s == "" {
		return fmt.Errorf("string elements cannot be empty")
	}

	escaped, err := escapeXML(string(s))
	if err != nil {
		return fmt.Errorf("invalid string element: %w", err)
	}

	v := struct {
		Elem string `xml:",innerxml"`
	}{Elem: "<" + escaped + "/>"}
	return e.EncodeElement(&v, start)
}

type SentinalBool bool

func (b SentinalBool) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if !b {
		return nil
	}
	fmt.Println(b)

	return e.EncodeElement(b, start)
}

func (b *SentinalBool) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	v := &struct{}{}
	if err := d.DecodeElement(v, &start); err != nil {
		return err
	}
	*b = v != nil
	return nil
}

func escapeXML(input string) (string, error) {
	buf := &strings.Builder{}
	if err := xml.EscapeText(buf, []byte(input)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
