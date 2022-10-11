package ssh

import (
	"context"
	"fmt"
	"net"

	"github.com/nemith/netconf/transport"
	"golang.org/x/crypto/ssh"
)

// alias it to a private type so we can make it private when embedding
type framer = transport.Framer

// Transport implements RFC6242 for implementing NETCONF protocol over SSH.
type Transport struct {
	c    *ssh.Client
	sess *ssh.Session

	// indicate that we "own" the client and should close it with the session
	// when the transport is closed.
	ownedClient bool

	*framer
}

// Dial will connect to a ssh server and issues a transport, it's used as a
// convience function as essnetial is the same as
//
//		c, err := ssh.Dial(network, addr, config)
//	 	if err != nil { /* ... handle error ... */ }
//	 	t, err := NewTransport(c)
//
// When the transport is closed the underlying connection is also closed.
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
// are still closed).
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

		framer: transport.NewFramer(r, w),
	}, nil
}

// Close will close the underlying transport.  If the connection was created
// with Dial then then underlying ssh.Client is closed as well.  If not only
// the sessions is closed.
func (t *Transport) Close() error {
	if err := t.sess.Close(); err != nil {
		return err
	}

	if t.ownedClient {
		return t.c.Close()
	}

	return nil
}
