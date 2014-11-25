package netconf

import (
	"encoding/xml"
)

type Session struct {
	Transport          Transport
	SessionID          int
	ServerCapabilities []string
	ErrOnWarning       bool
}

func (s *Session) Close() error {
	return s.Transport.Close()
}

func (s *Session) Exec(methods ...RPCMethod) (*RPCReply, error) {
	rpc := NewRpcMessage(methods)

	request, err := xml.Marshal(rpc)
	if err != nil {
		return nil, err
	}

	log.Debugf("REQUEST: %s\n", request)

	err = s.Transport.Send(request)
	if err != nil {
		return nil, err
	}

	rawXML, err := s.Transport.Receive()
	if err != nil {
		return nil, err
	}
	log.Debugf("REPLY: %s\n", rawXML)

	reply := &RPCReply{}
	reply.RawReply = string(rawXML)

	if err := xml.Unmarshal(rawXML, reply); err != nil {
		return nil, err
	}

	if reply.Errors != nil {
		// We have errors, lets see if it's a warning or an error.
		for _, rpcErr := range reply.Errors {
			if rpcErr.Severity == "error" || s.ErrOnWarning {
				return reply, &rpcErr
			}
		}

	}

	return reply, nil
}

func NewSession(t Transport) *Session {
	s := new(Session)
	s.Transport = t

	// Receive Servers Hello message
	serverHello, _ := t.ReceiveHello()
	s.SessionID = serverHello.SessionID
	s.ServerCapabilities = serverHello.Capabilities

	// Send our hello using default capabilities.
	t.SendHello(&HelloMessage{Capabilities: DEFAULT_CAPABILITIES})

	return s
}
