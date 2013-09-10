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

func (s *Session) ExecRPC(rpc *RPCMessage) (*RPCReply, error) {
	val, err := xml.MarshalIndent(rpc, "  ", "    ")
	if err != nil {
		return nil, err
	}

	return s.Exec(val)
}

func (s *Session) Exec(msg []byte) (*RPCReply, error) {
	reply := new(RPCReply)

	err := s.Transport.Send(msg)
	if err != nil {
		return reply, err
	}

	rawReply, err := s.Transport.Receive()
	if err != nil {
		return reply, err
	}

	reply.XML = rawReply

	if err := xml.Unmarshal([]byte(rawReply), reply); err != nil {
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
