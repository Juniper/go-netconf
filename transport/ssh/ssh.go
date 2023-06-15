package ssh

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/nemith/netconf/transport"
	"golang.org/x/crypto/ssh"
)

// alias it to a private type so we can make it private when embedding
type framer = transport.Framer //nolint:golint,unused

// Transport implements RFC6242 for implementing NETCONF protocol over SSH.
type Transport struct {
	c     *ssh.Client
	sess  *ssh.Session
	stdin io.WriteCloser

	// set to true if the transport is managing the underlying ssh connection
	// and should close it when the transport is closed.  This is is set to true
	// when used with `Dial`.
	managed bool

	*framer
}

// Dial will connect to a ssh server and issues a transport, it's used as a
// convenience function as essentially is the same as
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

	// Setup a go routine to monitor the context and close the connection.  This
	// is needed as the underlying ssh library doesn't support contexts so this
	// approximates a context based cancelation/timeout for the ssh handshake.
	//
	// An alternative would be timeout based with conn.SetDeadline(), but then we
	// would manage two timeouts.  One for tcp connection and one for ssh
	// handshake and wouldn't support any other event based cancelation.
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			// context is canceled so close the underlying connection.  Will
			// will catch ctx.Err() later.
			conn.Close()
		case <-done:
		}
	}()

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		// if there is a context timeout return that error instead of the actual
		// error from ssh.NewClientConn.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	close(done) // make sure we cleanup the context monitor routine

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

func newTransport(client *ssh.Client, managed bool) (*Transport, error) {
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
		c:       client,
		managed: managed,
		sess:    sess,
		stdin:   w,

		framer: transport.NewFramer(r, w),
	}, nil
}

// Close will close the underlying transport.  If the connection was created
// with Dial then then underlying ssh.Client is closed as well.  If not only
// the sessions is closed.
func (t *Transport) Close() error {
	// TODO: in go 1.20 this could easily be an errors.Join() but for now we
	// will save previous errors but try to close everything returning just the
	// "lowest" abstraction layer error
	var retErr error

	if err := t.stdin.Close(); err != nil {
		retErr = fmt.Errorf("failed to close ssh stdin: %w", err)
	}

	if err := t.sess.Close(); err != nil {
		retErr = fmt.Errorf("failed to close ssh channel: %w", err)
	}

	if t.managed {
		if err := t.c.Close(); err != nil {
			return fmt.Errorf("failed to close ssh connnection: %w", t.c.Close())
		}
	}

	return retErr
}
