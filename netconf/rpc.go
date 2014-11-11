package netconf

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type RPCMessage struct {
	MessageId string
	Methods   []RPCMethod
}

func NewRpcMessage(methods []RPCMethod) *RPCMessage {
	return &RPCMessage{
		MessageId: uuid(),
		Methods:   methods,
	}
}

func (m *RPCMessage) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var buf bytes.Buffer
	for _, method := range m.Methods {
		buf.WriteString(method.MarshalMethod())
	}

	data := struct {
		MessageId string `xml:"message-id,attr"`
		Methods   []byte `xml:",innerxml"`
	}{
		m.MessageId,
		buf.Bytes(),
	}

	start.Name.Local = "rpc"
	return e.EncodeElement(data, start)
}

type RPCReply struct {
	XMLName  xml.Name   `xml:"rpc-reply"`
	Errors   []RPCError `xml:"rpc-error,omitempty"`
	Data     string     `xml:",innerxml"`
	Ok       bool       `xml:",omitempty"`
	RawReply string     `xml:"-"`
}

type RPCError struct {
	Type     string `xml:"error-type"`
	Tag      string `xml:"error-tag"`
	Severity string `xml:"error-severity"`
	Path     string `xml:"error-path"`
	Message  string `xml:"error-message"`
	Info     string `xml:",innerxml"`
}

func (re *RPCError) Error() string {
	return fmt.Sprintf("netconf rpc [%s] '%s'", re.Severity, re.Message)
}

type RPCMethod interface {
	MarshalMethod() string
}

type RawMethod string

func (r RawMethod) MarshalMethod() string {
	return string(r)
}

func MethodLock(target string) RawMethod {
	return RawMethod(fmt.Sprintf("<lock><target><%s/></target></lock>", target))
}

func MethodUnlock(target string) RawMethod {
	return RawMethod(fmt.Sprintf("<unlock><target><%s/></target></unlock>", target))
}

func MethodGetConfig(source string) RawMethod {
	return RawMethod(fmt.Sprintf("<get-config><source><%s/></source></get-config>", source))
}
