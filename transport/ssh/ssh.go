package ssh

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"

	"github.com/nemith/go-netconf/v2/transport"
	"golang.org/x/crypto/ssh"
)

type Transport struct {
	c    *ssh.Client
	sess *ssh.Session
	r    *bufio.Reader
	w    *bufio.Writer

	// indicate that we "own" the client
	ownedClient bool
	upgraded    bool
}

func Dial(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*Transport, error) {
	d := net.Dialer{Timeout: config.Timeout}
	conn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	return newTransport(client, true)
}

func NewTransport(client *ssh.Client) (*Transport, error) {
	return newTransport(client, false)
}

func newTransport(client *ssh.Client, owned bool) (*Transport, error) {
	sess, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh session: %w", err)
	}

	w, err := sess.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	r, err := sess.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	const subsystem = "netconf"
	if err := sess.RequestSubsystem(subsystem); err != nil {
		return nil, fmt.Errorf("failed to start netconf ssh subsytem: %w", err)
	}

	return &Transport{
		c:           client,
		ownedClient: owned,
		sess:        sess,
		r:           bufio.NewReader(r),
		w:           bufio.NewWriter(w),
	}, nil
}

func (t *Transport) MsgWriter() io.WriteCloser {
	if t.upgraded {
		return transport.NewChunkWriter(t.w)
	}
	return transport.NewFrameWriter(t.w)
}

func (t *Transport) MsgReader() io.Reader {
	if t.upgraded {
		return transport.NewChunkReader(t.r)
	}
	return transport.NewFrameReader(t.r)
}

func (t *Transport) Close() error {
	if err := t.sess.Close(); err != nil {
		return err
	}

	if t.ownedClient {
		return t.c.Close()
	}
	return nil
}
