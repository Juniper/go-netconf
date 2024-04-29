// Go NETCONF Client
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netconf

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	// bastionDefaultPort is the default port used when SSH via Bastion
	bastionDefaultPort = 22
	// sshDefaultPort is the default SSH port used when communicating with
	// NETCONF
	sshDefaultPort = 830
	// sshNetconfSubsystem sets the SSH subsystem to NETCONF
	sshNetconfSubsystem = "netconf"
)

// TransportSSH maintains the information necessary to communicate with the
// remote device over SSH
type TransportSSH struct {
	TransportBasicIO
	sshClient  *ssh.Client
	sshSession *ssh.Session
}

// Close closes an existing SSH session and socket if they exist.
func (t *TransportSSH) Close() error {
	// If TransportSSH is nil ignore closing ssh session
	if t == nil {
		return nil
	}

	// Close the SSH Session if we have one
	if t.sshSession != nil {
		if err := t.sshSession.Close(); err != nil {
			// If we receive an error when trying to close the session, then
			// lets try to close the socket, otherwise it will be left open
			t.sshClient.Close()
			return err
		}
	}

	// Close the socket
	if t.sshClient != nil {
		return t.sshClient.Close()
	}
	return fmt.Errorf("No connection to close")
}

// DialWithBasstion connects and establishes SSH sessions over bastion host
//
// bastion can be an IP address (e.g) 172.16.2.1 which uses default port 22.
// Bastion can also specify port with format <host>:<port> (e.g 172.16.2.1:5001)
// target can be an IP address (e.g.) 172.16.1.1 which utlizes the default
// NETCONF over SSH port of 830.  Target can also specify a port with the
// following format <host>:<port (e.g 172.16.1.1:22)
//
// config takes a ssh.ClientConfig connection. See documentation for
// go.crypto/ssh for documenation.  There is a helper function SSHConfigPassword
// thar returns a ssh.ClientConfig for simple username/password authentication
func (t *TransportSSH) DialWithBasstion(bastion, target string, bastionConf, config *ssh.ClientConfig) error {
	var err error

	if !strings.Contains(bastion, ":") {
		bastion = fmt.Sprintf("%s:%d", bastion, bastionDefaultPort)
	}

	bastionClient, err := ssh.Dial("tcp", bastion, bastionConf)
	if err != nil {
		return err
	}

	if !strings.Contains(target, ":") {
		target = fmt.Sprintf("%s:%d", target, sshDefaultPort)
	}

	targetConn, err := bastionClient.Dial("tcp", target)
	if err != nil {
		return err
	}

	ncc, chans, reqs, err := ssh.NewClientConn(targetConn, target, config)
	if err != nil {
		return err
	}

	t.sshClient = ssh.NewClient(ncc, chans, reqs)
	if t.sshClient == nil {
		return errors.New("Error getting SSH Client for " + target + " via Bastion " + bastion)
	}

	err = t.setupSession()
	return err
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
	return err
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
	return t.sshSession.RequestSubsystem(sshNetconfSubsystem)
}

// NewSSHSession creates a new NETCONF session using an existing net.Conn.
func NewSSHSession(conn net.Conn, config *ssh.ClientConfig) (*Session, error) {
	t, err := connToTransport(conn, config)
	if err != nil {
		return nil, err
	}

	return NewSession(t), nil
}

// DialSSHWithBastion creates a new NETCONF session using a SSH Transport.
// See TransportSSH.Dial for arguments.
func DialSSHWithBastion(bastion, target string, bastionConf, config *ssh.ClientConfig) (*Session, error) {
	var t TransportSSH
	err := t.DialWithBasstion(bastion, target, bastionConf, config)
	if err != nil {
		t.Close()
		return nil, err
	}
	return NewSession(&t), nil
}

// DialSSH creates a new NETCONF session using a SSH Transport.
// See TransportSSH.Dial for arguments.
func DialSSH(target string, config *ssh.ClientConfig) (*Session, error) {
	var t TransportSSH
	err := t.Dial(target, config)
	if err != nil {
		t.Close()
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
		if t != nil {
			t.Close()
		}
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

// SSHConfigPassword is a convenience function that takes a username and password
// and returns a new ssh.ClientConfig setup to pass that username and password.
// Convenience means that HostKey checks are disabled so it's probably less secure
func SSHConfigPassword(user string, pass string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

// SSHConfigPubKeyFile is a convenience function that takes a username, private key
// and passphrase and returns a new ssh.ClientConfig setup to pass credentials
// to DialSSH
func SSHConfigPubKeyFile(user string, file string, passphrase string) (*ssh.ClientConfig, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	block, rest := pem.Decode(buf)
	if len(rest) > 0 {
		return nil, fmt.Errorf("pem: unable to decode file %s", file)
	}

	if x509.IsEncryptedPEMBlock(block) {
		b, err := x509.DecryptPEMBlock(block, []byte(passphrase))
		if err != nil {
			return nil, err
		}
		buf = pem.EncodeToMemory(&pem.Block{
			Type:  block.Type,
			Bytes: b,
		})
	}

	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}, nil

}

// SSHConfigPubKeyAgent is a convience function that takes a username and
// returns a new ssh.Clientconfig setup to pass credentials received from
// an ssh agent
func SSHConfigPubKeyAgent(user string) (*ssh.ClientConfig, error) {
	c, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agent.NewClient(c).Signers),
		},
	}, nil
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
