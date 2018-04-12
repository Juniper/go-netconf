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

// Close closes an existing session.
func (t *TransportJunOS) Close() error {
	return nil
}

// Open creates a new session.
func (t *TransportJunOS) Open() error {
	var err error

	err = t.setup()
	if err != nil {
		return err
	}
	return nil
}

// setup executes the local shell command.
func (t *TransportJunOS) setup() error {
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
