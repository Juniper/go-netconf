// Copyright (c) 2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netconf

import (
	"net"
	"os/exec"

	session "github.com/arsonistgopher/go-netconf/session"
	transport "github.com/arsonistgopher/go-netconf/transport"
	"golang.org/x/crypto/ssh"
)

// DriverSSH type is for creating an SSH based driver. Maintains state for session and connection. Implements Driver{}
type DriverSSH struct {
	Port      int               // Target port
	Target    string            // Target host
	Conn      net.Conn          // Conn for session
	SSHConfig *ssh.ClientConfig // SSH Config
	Transport *TransportJunos   // Transport data
}

// TransportJunos maintains the information necessary to communicate with Junos
// via local shell NETCONF interface.
type TransportJunos struct {
	transport.TransportBasicIO
	cmd *exec.Cmd
}

// Close closes an existing local NETCONF session.
func (t *TransportJunos) Close() error {
	if t.cmd != nil {
		t.ReadWriteCloser.Close()
	}
	return nil
}

// Open creates a new local NETCONF session.
func (t *TransportJunos) Open() error {
	var err error

	t.cmd = exec.Command("xml-mode", "netconf", "need-trailer")

	w, err := t.cmd.StdinPipe()
	if err != nil {
		return err
	}

	r, err := t.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	t.ReadWriteCloser = transport.NewReadWriteCloser(r, w)
	return t.cmd.Start()
}

// Dial creates a new NETCONF session via Junos local shell
// NETCONF interface (xml-mode netconf need-trailer).
func Dial() (*session.Session, error) {
	var t TransportJunos
	err := t.Open()
	if err != nil {
		return nil, err
	}
	return session.NewSession(&t), nil
}
