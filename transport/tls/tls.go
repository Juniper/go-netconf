package tls

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/nemith/netconf/transport"
)

// alias it to a private type so we can make it private when embedding
type framer = transport.Framer //nolint:golint,unused

// Transport implements RFC7589 for implementing NETCONF over TLS.
type Transport struct {
	conn *tls.Conn
	*framer
}

// Dial will connect to a server via TLS and retuns a Transport.
func Dial(ctx context.Context, network, addr string, config *tls.Config) (*Transport, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	tlsConn := tls.Client(conn, config)
	return NewTransport(tlsConn), nil

}

// NewTransport takes an already connected tls transport and returns a new
// Transport.
func NewTransport(conn *tls.Conn) *Transport {
	return &Transport{
		conn:   conn,
		framer: transport.NewFramer(conn, conn),
	}
}

// Close will close the transport and the underlying TLS connection.
func (t *Transport) Close() error {
	return t.conn.Close()
}
