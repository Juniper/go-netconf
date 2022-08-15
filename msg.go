package netconf

import (
	"context"
	"encoding/xml"
	"time"
)

// helloMsg maps the xml value of the <hello> message in RFC6241
type HelloMsg struct {
	XMLName      xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 hello"`
	SessionID    uint64   `xml:"session-id,omitempty"`
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
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc-reply"`
	MessageID uint64   `xml:"message-id,attr"`

	// Ok is part of RFC6241 and is present if no data is returned from an
	// RPC call and there were no errors.  This IS NOT set to true if data is
	// also returned.  To check if a call is ok then look ath the RPCErrors

	Errors []RPCError `xml:"rpc-error,omitempty"`
	Data   []byte     `xml:",innerxml"`
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

// XXX: RPC calls these either Methods or Operations depending on what you look at.
type GetConfigRPC struct {
	Source StringElem
	Filter Filter
}

type GetConfigResp struct {
}

func (s *Session) GetConfig(ctx context.Context, source string) ([]byte, error) {
	method := struct {
		// XXX do these need namespaced as well?
		XMLName xml.Name   `xml:"get-config"`
		Source  StringElem `xml:"source"`
		// Filter
	}{Source: StringElem(source)}

	resp := struct {
		// XXX do these need namespaced as well?
		XMLName xml.Name `xml:"data"`
		Config  []byte   `xml:",innerxml"`
	}{}

	if err := s.Call(ctx, &method, &resp); err != nil {
		return nil, err
	}

	return resp.Config, nil
}

type OKResponse struct {
	Ok SentinalBool `xml:"ok"`
}

// <get-config>
//    source, filter
//
// <edit-config>
//    operation,

// <copy-config>
// <delete-config>
// <lock>
// <unlock>
// <get>
// <close>  // already implemented and hidden...
// <kill-session>
