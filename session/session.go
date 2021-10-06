// Go NETCONF Client
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package netconf is a simple NETCONF client based on RFC6241 and RFC6242
(although not fully compliant yet).
*/
package netconf

import (
	"encoding/xml"

	rpc "github.com/davedotdev/go-netconf/rpc"
	transport "github.com/davedotdev/go-netconf/transport"
)

// Session defines the necessary components for a NETCONF session
type Session struct {
	Transport          transport.Transport
	SessionID          int
	ServerCapabilities []string
	ErrOnWarning       bool
}

// Close is used to close and end a transport session
func (s *Session) Close() error {
	return s.Transport.Close()
}

// Exec is used to execute an RPC method or methods
func (s *Session) Exec(methods ...rpc.RPCMethod) (*rpc.RPCReply, error) {
	rpcm := rpc.NewRPCMessage(methods)

	request, err := xml.Marshal(rpcm)
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

	reply, err := rpc.NewRPCReply(rawXML, s.ErrOnWarning)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

// NewSession creates a new NETCONF session using the provided transport layer.
func NewSession(t transport.Transport) (*Session, error) {
	s := new(Session)
	s.Transport = t

	// Receive Servers Hello message
	serverHello, err := t.ReceiveHello()
	if err != nil {
		return nil, err
	}
	s.SessionID = serverHello.SessionID
	s.ServerCapabilities = serverHello.Capabilities

	// Send our hello using default capabilities.
	t.SendHello(&transport.HelloMessage{Capabilities: transport.DefaultCapabilities})

	return s, nil
}
