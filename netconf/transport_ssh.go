package netconf

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"strings"
)

const (
	SSH_DEFAULT_PORT      = 830
	SSH_NETCONF_SUBSYSTEM = "netconf"
)

type TransportSSH struct {
	transportBasicIO
	sshConn    *ssh.ClientConn
	sshSession *ssh.Session
}

func (t *TransportSSH) Close() error {
	// Close the SSH Session if we have one
	if t.sshSession != nil {
		if err := t.sshSession.Close(); err != nil {
			return err
		}
	}

	// Close the socket
	if err := t.sshConn.Close(); err != nil {
		return err
	}

	return nil
}

// Dials and establishes an SSH sessions
//
// target can be an IP address (e.g.) 172.16.1.1 which utlizes the default
// NETCONF over SSH port of 830.  Target can also specify a port with the
// following format <host>:<port (e.g 172.16.1.1:22)
//
// config takes a ssh.ClientConfig connection. See documentation for
// go.crypto/ssh for documenation.  There is a helper function SSHConfigPassword
// thar returns a ssh.ClientConfig for simple username/password authentication
func (t *TransportSSH) Dial(target string, config *ssh.ClientConfig) error {
	if !strings.Contains(target, ":") {
		target = fmt.Sprintf("%s:%d", target, SSH_DEFAULT_PORT)
	}

	var err error

	t.sshConn, err = ssh.Dial("tcp", target, config)
	if err != nil {
		return err
	}

	t.sshSession, err = t.sshConn.NewSession()
	if err != nil {
		return err
	}

	writer, err := t.sshSession.StdinPipe()
	if err != nil {
		return err
	}

	reader, err := t.sshSession.StdoutPipe()
	if err != nil {
		return err
	}

	t.io = NewReadWriteCloser(reader, writer)

	if err := t.sshSession.RequestSubsystem(SSH_NETCONF_SUBSYSTEM); err != nil {
		return err
	}

	return nil
}

// Create a new NETCONF session using a SSH Transport. See TransportSSH.Dial for arguments.
func DialSSH(target string, config *ssh.ClientConfig) (*Session, error) {
	var t TransportSSH
	err := t.Dial(target, config)
	if err != nil {
		return nil, err
	}
	return NewSession(&t), nil
}

type simpleSSHPassword string

func (p simpleSSHPassword) Password(user string) (string, error) {
	return string(p), nil
}

// SSHConfigPassword is a convience function that takes a username and password
// and returns a new ssh.ClientConfig setup to pass that username and password.
func SSHConfigPassword(user string, pass string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(simpleSSHPassword(pass)),
		},
	}
}
