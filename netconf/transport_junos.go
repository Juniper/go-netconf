// Copyright (c) 2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Transport JunOS provides the ability to communicate with JunOS via local shell
NETCONF interface (xml-mode netconf need-trailer).
*/
package netconf

import (
	"os/exec"
)

// TransportJunOS maintains the information necessary to communicate with JunOS
// via local shell.
type TransportJunOS struct {
	transportBasicIO
	cmd *exec.Cmd
}

// Close closes an existing local NETCONF session.
func (t *TransportJunOS) Close() error {
	if t.cmd != nil {
		t.ReadWriteCloser.Close()
	}
	return nil
}

// Open creates a new local NETCONF session.
func (t *TransportJunOS) Open() error {
	var err error

	t.cmd = exec.Command("xml-mode", "netconf", "need-trailer")

	writer, err := t.cmd.StdinPipe()
	if err != nil {
		return err
	}

	reader, err := t.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	t.ReadWriteCloser = NewReadWriteCloser(reader, writer)
	return t.cmd.Start()
}

// JunOS creates a new NETCONF session using shell transport.
func JunOS() (*Session, error) {
	var t TransportJunOS
	err := t.Open()
	if err != nil {
		return nil, err
	}
	return NewSession(&t), nil
}
