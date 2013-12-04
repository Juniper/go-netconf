package netconf

import (
	"github.com/tarm/goserial"
)

type TransportSerial struct {
	transportBasicIO
}

func (t *TransportSerial) Dial(port string, baud int, username string, password string, vendor VendorIOProc) error {
	serialConfig := &serial.Config{Name: port, Baud: baud}
	s, err := serial.OpenPort(serialConfig)
	if err != nil {
		return err
	}

	t.ReadWriteCloser = s
	t.chunkedFraming = false

	// Send a new line to get the prompt to show up
	t.Write([]byte("\n\n"))
	vendor.Login(t, username, password)
	vendor.StartNetconf(t)

	return nil
}

func DialSerial(port string, baud int, username string, password string, vendor VendorIOProc) (*Session, error) {
	var t TransportSerial
	if err := t.Dial(port, baud, username, password, vendor); err != nil {
		return nil, err
	}
	return NewSession(&t), nil
}
