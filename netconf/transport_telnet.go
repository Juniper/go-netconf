package netconf

import (
	"fmt"
	"strings"
	"time"

	"github.com/ziutek/telnet"
)

const (
	// telnetDefaultPort sets the default port for use by Telnet
	telnetDefaultPort = 23
	// telnetTimeout sets the timeout duration for use by Telnet
	telnetTimeout = 10 * time.Second
)

// VendorIOProc is the interface used when establishing a telnet NETCONF session
type VendorIOProc interface {
	Login(*TransportTelnet, string, string) error
	StartNetconf(*TransportTelnet) error
}

// TransportTelnet is used to define what makes up a Telnet Transport layer for
// NETCONF
type TransportTelnet struct {
	transportBasicIO
	telnetConn *telnet.Conn
}

// Dial is used to create a TCP Telnet connection to the remote host returning
// only an error if it is unable to dial the remote host.
func (t *TransportTelnet) Dial(target string, username string, password string, vendor VendorIOProc) error {
	if !strings.Contains(target, ":") {
		target = fmt.Sprintf("%s:%d", target, telnetDefaultPort)
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

// DialTelnet dials and returns the usable telnet session.
func DialTelnet(target string, username string, password string, vendor VendorIOProc) (*Session, error) {
	var t TransportTelnet
	if err := t.Dial(target, username, password, vendor); err != nil {
		return nil, err
	}
	return NewSession(&t), nil
}
