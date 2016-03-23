package netconf

import (
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	// sshDefaultPort is the default SSH port used when communicating with NetConf
	sshDefaultPort = 830
	// sshNetconfSubsystem sets the SSH subsystem to NetConf
	sshNetconfSubsystem = "netconf"
)

// TransportSSH maintains the information necessary to communicate with the
// remote device over SSH
type TransportSSH struct {
	transportBasicIO
	sshClient  *ssh.Client
	sshSession *ssh.Session
}

// Close closes an existing SSH session and socket if they exist.
func (t *TransportSSH) Close() error {
	// Close the SSH Session if we have one
	if t.sshSession != nil {
		if err := t.sshSession.Close(); err != nil {
			return err
		}
	}

	// Close the socket
	if err := t.sshClient.Close(); err != nil {
		return err
	}

	return nil
}

// Dial connects and establishes SSH sessions
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
		target = fmt.Sprintf("%s:%d", target, sshDefaultPort)
	}

	var err error

	t.sshClient, err = ssh.Dial("tcp", target, config)
	if err != nil {
		return err
	}

	err = t.setupSession()
	if err != nil {
		return err
	}

	return nil
}

func (t *TransportSSH) setupSession() error {
	var err error

	t.sshSession, err = t.sshClient.NewSession()
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

	t.ReadWriteCloser = NewReadWriteCloser(reader, writer)

	if err := t.sshSession.RequestSubsystem(sshNetconfSubsystem); err != nil {
		return err
	}

	return nil
}

// NewSSHSession creates a new NETCONF session using an existing net.Conn.
func NewSSHSession(conn net.Conn, config *ssh.ClientConfig) (*Session, error) {
	t, err := connToTransport(conn, config)
	if err != nil {
		return nil, err
	}

	return NewSession(t), nil
}

// DialSSH creates a new NETCONF session using a SSH Transport.
// See TransportSSH.Dial for arguments.
func DialSSH(target string, config *ssh.ClientConfig) (*Session, error) {
	var t TransportSSH
	err := t.Dial(target, config)
	if err != nil {
		return nil, err
	}
	return NewSession(&t), nil
}

// DialSSHTimeout creates a new NETCONF session using a SSH Transport with timeout.
// See TransportSSH.Dial for arguments.
// The timeout value is used for both connection establishment and Read/Write operations.
func DialSSHTimeout(target string, config *ssh.ClientConfig, timeout time.Duration) (*Session, error) {
	bareConn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return nil, err
	}

	conn := &deadlineConn{Conn: bareConn, timeout: timeout}
	t, err := connToTransport(conn, config)
	if err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(timeout / 2)
		defer ticker.Stop()
		for range ticker.C {
			_, _, err := t.sshClient.Conn.SendRequest("KEEP_ALIVE", true, nil)
			if err != nil {
				return
			}
		}
	}()

	return NewSession(t), nil
}

// SSHConfigPassword is a convience function that takes a username and password
// and returns a new ssh.ClientConfig setup to pass that username and password.
func SSHConfigPassword(user string, pass string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
	}
}

func connToTransport(conn net.Conn, config *ssh.ClientConfig) (*TransportSSH, error) {
	c, chans, reqs, err := ssh.NewClientConn(conn, conn.RemoteAddr().String(), config)
	if err != nil {
		return nil, err
	}

	t := &TransportSSH{}
	t.sshClient = ssh.NewClient(c, chans, reqs)

	err = t.setupSession()
	if err != nil {
		return nil, err
	}

	return t, nil
}

type deadlineConn struct {
	net.Conn
	timeout time.Duration
}

func (c *deadlineConn) Read(b []byte) (n int, err error) {
	c.SetReadDeadline(time.Now().Add(c.timeout))
	return c.Conn.Read(b)
}

func (c *deadlineConn) Write(b []byte) (n int, err error) {
	c.SetWriteDeadline(time.Now().Add(c.timeout))
	return c.Conn.Write(b)
}
