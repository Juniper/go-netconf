package tls

import (
	"context"
	"crypto/tls"
	"io"
	"net"

	"github.com/nemith/go-netconf/v2/transport"
)

type Transport struct {
	conn   *tls.Conn
	framer transport.Transport
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
		conn:   conn,
		framer: transport.NewFrameTransport(conn, conn),
	}
}

func (t *Transport) MsgReader() (io.Reader, error)      { return t.framer.MsgReader() }
func (t *Transport) MsgWriter() (io.WriteCloser, error) { return t.framer.MsgWriter() }

func (t *Transport) Close() error {
	t.framer.Close()
	return t.conn.Close()
}
