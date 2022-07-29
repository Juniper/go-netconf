package tls

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/nemith/go-netconf/v2/transport"
)

type Transport struct {
	conn *tls.Conn
	*transport.FramedTransport
}

func Dial(ctx context.Context, network, addr string, config *tls.Config) (*Transport, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	tlsConn := tls.Client(conn, config)
	return NewTransport(tlsConn), nil

}

func NewTransport(conn *tls.Conn) *Transport {
	return &Transport{
		conn:            conn,
		FramedTransport: transport.NewFramedTransport(conn, conn),
	}
}

func (c *Transport) Close() error {
	return c.conn.Close()
}
