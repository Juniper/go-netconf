package netconf

import (
	"fmt"
	"net"
	"time"

	lowlevel "github.com/arsonistgopher/go-netconf/drivers/ssh/lowlevel"
	rpc "github.com/arsonistgopher/go-netconf/rpc"
	session "github.com/arsonistgopher/go-netconf/session"
	"golang.org/x/crypto/ssh"
)

/* Type below implements this interface
// Driver interface for building drivers that are self-contained from a user's perspective.
type Driver interface {
	SetDatastore(string) error
	Lock() (*rpc.RPCReply, error)
	Unlock() (*rpc.RPCReply, error)
	Close() error
	Dial() error
	DialTimeout() error
	SendRaw() (*rpc.RPCReply, error)
	GetConfig() (*rpc.RPCReply, error)
}
*/

// DriverSSH type is for creating an SSH based driver. Maintains state for session and connection. Implements Driver{}
type DriverSSH struct {
	Timeout   time.Duration          // Timeout for SSH timed sessions
	Port      int                    // Target port
	Host      string                 // Target hostname
	Target    string                 // Target hostname:port
	Datastore string                 // NETCONF datastore
	Conn      net.Conn               // Conn for session
	SSHConfig *ssh.ClientConfig      // SSH Config
	Transport *lowlevel.TransportSSH // Transport data
	Session   *session.Session       // Session data
}

// New creates a new instance of DriverSSH
func New() *DriverSSH {
	var t lowlevel.TransportSSH
	return &DriverSSH{
		Port:      lowlevel.DefaultPort,
		Transport: &t,
	}
}

// SetDatastore sets the target datastore on the data structure
func (d *DriverSSH) SetDatastore(ds string) error {
	d.Datastore = ds
	return nil
}

// Dial function (call this after New())
func (d *DriverSSH) Dial() error {
	d.Target = fmt.Sprintf("%s:%d", d.Host, d.Port)

	var err error

	err = d.Transport.DialSSH(d.Host, d.SSHConfig, d.Port)
	d.Session, err = lowlevel.Dial(d.Host, d.SSHConfig, d.Port)

	if err != nil {
		return err
	}

	// err = d.Transport.SetupSession()

	if err != nil {
		return err
	}

	return nil
}

// DialTimeout function (call this after New())
func (d *DriverSSH) DialTimeout() error {
	d.Target = fmt.Sprintf("%s:%d", d.Host, d.Port)

	var err error

	d.Session, err = lowlevel.DialSSHTimeout(d.Target, d.SSHConfig, d.Timeout)

	if err != nil {
		return err
	}

	err = d.Transport.SetupSession()

	if err != nil {
		return err
	}

	return nil
}

// Close function closes the socket
func (d *DriverSSH) Close() error {
	// Close the SSH Session if we have one}

	err := d.Transport.Close()

	if err != nil {
		return err
	}

	return nil
}

// Lock the target datastore
func (d *DriverSSH) Lock() (*rpc.RPCReply, error) {
	reply, err := d.Session.Exec(rpc.MethodLock(d.Datastore))

	if err != nil {
		return reply, err
	}

	return reply, nil
}

// Unlock the target datastore
func (d *DriverSSH) Unlock() (*rpc.RPCReply, error) {
	reply, err := d.Session.Exec(rpc.MethodUnlock(d.Datastore))

	if err != nil {
		return reply, err
	}

	return reply, nil
}

// SendRaw sends a raw XML envelope
func (d *DriverSSH) SendRaw(rawxml string) (*rpc.RPCReply, error) {
	reply, err := d.Session.Exec(rpc.RawMethod(rawxml))

	if err != nil {
		return reply, err
	}

	return reply, nil
}

// GetConfig requests the contents of a datastore
func (d *DriverSSH) GetConfig() (*rpc.RPCReply, error) {
	reply, err := d.Session.Exec(rpc.MethodGetConfig(d.Datastore))

	if err != nil {
		return reply, err
	}

	return reply, nil
}
