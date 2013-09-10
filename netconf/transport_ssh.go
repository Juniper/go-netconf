package netconf

import (
	"bufio"
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

func splitOnSeperator(data []byte, atEOF bool) (advance int, token []byte, err error) {
	//fmt.Printf("splitOnSeperator(): Input data: '%s'\n", data)

	end := bytes.Index(data, []byte(MSG_SEPERATOR))
	//fmt.Printf("splitOnSeperator(): Index of seperator: '%d'\n", end)

	if end > 0 {
		//fmt.Printf("splitOnSeperator(): Found seperator.  Returning '%s'\n", data[:end])
		return end + len(MSG_SEPERATOR), data[:end], nil
	}

	//fmt.Printf("splitOnSeperator(): Requsting more data\n")
	return 0, nil, nil
}

func (t *TransportSSH) Receive() ([]byte, error) {
	scanner := bufio.NewScanner(t.sshStdout)
	scanner.Split(splitOnSeperator)

	scanner.Scan()

	return []byte(scanner.Text()), scanner.Err()
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
