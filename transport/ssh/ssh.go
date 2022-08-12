package ssh

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/nemith/go-netconf/v2/transport"
	"golang.org/x/crypto/ssh"
)

// Transport implements RFC6242 for implementing NETCONF protocol over ssh.
type Transport struct {
	c    *ssh.Client
	sess *ssh.Session
	// indicate that we "own" the client and should close it and the session.
	ownedClient bool

	framer transport.Transport
}

// Dial will connect to a ssh server and issues a transport, it's used as a
// convience function as essnetial is the same as
//
// 		c, err := ssh.Dial(networkm addrm config)
//  	if err != nil { /* ... handle error ... */ }
//  	t, err := NewTransport(c)
//
// When the transport is closed the ssh.Client is also closed.
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

// NewTransport will create a new ssh transport as defined in RFC6242 for use
// with netconf.  Unlike Dial, the underlying client will not be automatically
// closed when the transport is closed (however any sessions and subsystems
// are)
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

		framer: transport.NewFrameTransport(r, w),
	}, nil
}

func (t *Transport) MsgReader() (io.Reader, error)      { return t.framer.MsgReader() }
func (t *Transport) MsgWriter() (io.WriteCloser, error) { return t.framer.MsgWriter() }

// Close will close the underlying transport.  If the connection was created
// with Dial then then underlying ssh.Client is closed as well.  If not only
// the sessions is closed.
func (t *Transport) Close() error {
	t.framer.Close()

	if err := t.sess.Close(); err != nil {
		return err
	}

	if t.ownedClient {
		return t.c.Close()
	}

	return nil
}
