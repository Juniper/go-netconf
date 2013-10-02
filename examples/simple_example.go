package main

import (
	"code.google.com/p/go.crypto/ssh"
	"flag"
	"fmt"
	"github.com/Juniper/go-netconf/netconf"
	"os"
)

type clientPassword string

func (p clientPassword) Password(user string) (string, error) {
	return string(p), nil
}

func usage() {

}

func main() {
	username := flag.String("username", "", "User to login with")
	password := flag.String("password", "", "Password to login with")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] targets...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
	}

	sshConfig := &ssh.ClientConfig{
		User: *username,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(clientPassword(*password)),
		},
	}

	s, err := netconf.NewSessionSSH(flag.Arg(0), sshConfig)
	if err != nil {
		panic(err)
	}

	defer s.Close()

	fmt.Printf("Server Capabilities: '%+v'\n", s.ServerCapabilities)
	fmt.Printf("Session Id: %d\n\n", s.SessionID)

	//reply, err := s.Exec([]byte("<rpc><get-config><source><running/></source></get-config></rpc>"))
	reply, err := s.ExecRPC(netconf.RPCGetConfig("running"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply: %+v", reply)

}
