package netconf

import (
	//"bufio"
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const (
	SSH_DEFAULT_PORT      = 830
	SSH_NETCONF_SUBSYSTEM = "netconf"
)

type TransportSSH struct {
	Transport
	sshConn    *ssh.ClientConn
	sshSession *ssh.Session
	sshStdin   io.WriteCloser
	sshStdout  io.Reader
}

func (t *TransportSSH) Send(data []byte) error {
	t.sshStdin.Write(data)
	t.sshStdin.Write([]byte(MSG_SEPERATOR))
	t.sshStdin.Write([]byte("\n"))
	return nil // TODO: Implement error handling!
}

func (t *TransportSSH) Receive() ([]byte, error) {
	var out bytes.Buffer
	buf := make([]byte, 4096)

	for {
		n, err := t.sshStdout.Read(buf)

		if n == 0 {
			break // TODO: Handle Error
		}

		if err != nil {
			// TODO: Handle Error
			if err != io.EOF {
				fmt.Printf("Read error: %s", err)
			}
			break
		}

		end := bytes.Index(buf, []byte(MSG_SEPERATOR))
		if end > -1 {
			out.Write(buf[0:end])
			return out.Bytes(), nil
		}
		out.Write(buf[0:n])
	}

	return nil, nil
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

func (t *TransportSSH) SendHello(hello *HelloMessage) error {
	val, err := xml.MarshalIndent(hello, "  ", "    ")
	if err != nil {
		return err
	}

	err = t.Send(val)
	return err
}

func (t *TransportSSH) ReceiveHello() (*HelloMessage, error) {
	hello := new(HelloMessage)

	val, err := t.Receive()
	if err != nil {
		return hello, err
	}

	err = xml.Unmarshal([]byte(val), hello)
	return hello, err
}

func NewTranportSSH(target string, config *ssh.ClientConfig) (*TransportSSH, error) {
	if !strings.Contains(target, ":") {
		target = fmt.Sprintf("%s:%d", target, SSH_DEFAULT_PORT)
	}

	conn, err := ssh.Dial("tcp", target, config)
	if err != nil {
		return nil, err
	}

	sess, err := conn.NewSession()
	if err != nil {
		return nil, err
	}

	si, err := sess.StdinPipe()
	if err != nil {
		return nil, err
	}

	so, err := sess.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := sess.RequestSubsystem(SSH_NETCONF_SUBSYSTEM); err != nil {
		return nil, err
	}

	return &TransportSSH{sshConn: conn, sshSession: sess, sshStdin: si, sshStdout: so}, nil
}

func NewSessionSSH(target string, config *ssh.ClientConfig) (*Session, error) {
	t, err := NewTranportSSH(target, config)
	if err != nil {
		return nil, err
	}
	return NewSession(t), nil
}
