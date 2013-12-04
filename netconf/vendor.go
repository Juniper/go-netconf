package netconf

type VendorIOProc interface {
	Login(TransportIO, string, string) error
	StartNetconf(TransportIO) error
}
