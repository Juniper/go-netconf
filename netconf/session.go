package netconf

import (
	"encoding/xml"
)

// Session defines the necessary components for a NETCONF session
type Session struct {
	Transport          Transport
	SessionID          int
	ServerCapabilities []string
	ErrOnWarning       bool
}

// Close is used to close and end a transport session
func (s *Session) Close() error {
	return s.Transport.Close()
}

// Exec is used to execute an RPC method or methods
func (s *Session) Exec(methods ...RPCMethod) (*RPCReply, error) {
	rpc := NewRPCMessage(methods)

	request, err := xml.Marshal(rpc)
	if err != nil {
		return nil, err
	}

	header := []byte(xml.Header)
	request = append(header, request...)

	err = s.Transport.Send(request)
	if err != nil {
		return nil, err
	}

	rawXML, err := s.Transport.Receive()
	if err != nil {
		return nil, err
	}

	reply, err := NewRPCReply(rawXML, s.ErrOnWarning)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

// NewSession creates a new NETCONF session using the provided transport layer.
func NewSession(t Transport) *Session {
	s := new(Session)
	s.Transport = t

	// Receive Servers Hello message
	serverHello, _ := t.ReceiveHello()
	s.SessionID = serverHello.SessionID
	s.ServerCapabilities = serverHello.Capabilities

	// Send our hello using default capabilities.
	t.SendHello(&HelloMessage{Capabilities: DefaultCapabilities})

	return s
}
