package netconf

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"io"
)

func uuid() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type RPCMessage struct {
	XMLName    xml.Name    `xml:"rpc"`
	Message_id string      `xml:"message-id,attr,omitempty"`
	Operation  interface{} `xml:",innerxml"` // TODO: Support multiple operations
}

func NewRpcMessage(op interface{}) *RPCMessage {
	return NewRpcMessageID(uuid(), op)
}

func (rm *RPCMessage) String() string {
	val, err := xml.Marshal(rm)
	if err != nil {
		return ""
	}
	return string(val)
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
	return fmt.Sprintf("netconf: rpc %s: '%s'", re.Severity, re.Message)
}

func NewRpcMessageID(id string, op interface{}) *RPCMessage {
	return &RPCMessage{Message_id: id, Operation: op}
}

func RPCLock(target string) *RPCMessage {
	op := fmt.Sprintf("<lock><target><%s/></target></lock>", target)
	return NewRpcMessage(op)
}

func RPCUnlock(target string) *RPCMessage {
	op := fmt.Sprintf("<unlock><target><%s/></target></unlock>", target)
	return NewRpcMessage(op)
}

func RPCGetConfig(source string) *RPCMessage {
	op := fmt.Sprintf("<get-config><source><%s/></source></get-config>", source)
	return NewRpcMessage(op)
}
