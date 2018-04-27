// Go NETCONF Client
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package netconf provides a simple NETCONF client based on RFC6241 and RFC6242
(although not fully compliant yet).
*/
package netconf

import (
	"encoding/xml"
	"time"
)

// Session defines the necessary components for a NETCONF session.
type Session struct {
	Transport          Transport
	SessionID          int
	ServerCapabilities []string
	ErrOnWarning       bool
	RPCTimeout         time.Duration
}

// Close is used to close and end a transport session.
func (s *Session) Close() error {
	return s.Transport.Close()
}

// Exec is used to execute an RPC method or methods. It returns an
// RPCTimeout error if response is not received after Session.Timeout.
// A zero value for Session.Timeout means Exec operations will not time out.
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

	var rawXML []byte
	if s.RPCTimeout > 0 {
		c1 := make(chan ReceiveResult, 1)
		go func() {
			res := new(ReceiveResult)
			res.p, res.err = s.Transport.Receive()
			c1 <- *res
		}()
		select {
		case res := <-c1:
			rawXML, err = res.p, res.err
		case <-time.After(s.RPCTimeout):
			return nil, &RPCTimeoutError{s.RPCTimeout}
		}
	} else {
		rawXML, err = s.Transport.Receive()
	}
	if err != nil {
		return nil, err
	}

	reply, err := newRPCReply(rawXML, s.ErrOnWarning)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

// NewSession creates a new NETCONF session using the provided transport layer
// and RPC timeout value.
func NewSession(t Transport) *Session {
	s := new(Session)
	s.Transport = t

	// Receive server hello message
	serverHello, _ := t.ReceiveHello()
	s.SessionID = serverHello.SessionID
	s.ServerCapabilities = serverHello.Capabilities

	// Send client hello message using default capabilities
	t.SendHello(&HelloMessage{Capabilities: DefaultCapabilities})

	return s
}
