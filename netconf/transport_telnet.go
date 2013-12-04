package netconf

import (
	"fmt"
	"github.com/ziutek/telnet"
	"strings"
	"time"
)

const (
	TELNET_DEFAULT_PORT = 23
	TELNET_TIMEOUT      = 10 * time.Second
)

type VendorIOProc interface {
	Login(*TransportTelnet, string, string) error
	StartNetconf(*TransportTelnet) error
}

type TransportTelnet struct {
	transportBasicIO
	telnetConn *telnet.Conn
}

func (t *TransportTelnet) Dial(target string, username string, password string, vendor VendorIOProc) error {
	if !strings.Contains(target, ":") {
		target = fmt.Sprintf("%s:%d", target, TELNET_DEFAULT_PORT)
	}

	tn, err := telnet.Dial("tcp", target)
	if err != nil {
		return err
	}
	tn.SetUnixWriteMode(true)

	t.telnetConn = tn
	t.ReadWriteCloser = tn
	t.chunkedFraming = false

	vendor.Login(t, username, password)
	vendor.StartNetconf(t)

	return nil
}

func DialTelnet(target string, username string, password string, vendor VendorIOProc) (*Session, error) {
	var t TransportTelnet
	if err := t.Dial(target, username, password, vendor); err != nil {
		return nil, err
	}
	return NewSession(&t), nil
}
