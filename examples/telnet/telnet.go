package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Juniper/go-netconf/netconf"
	"github.com/Juniper/go-netconf/netconf/jnpr"
)

func main() {
	username := flag.String("username", "", "User to login with")
	password := flag.String("password", "", "Password to login with")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] targets...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
	}

	s, err := netconf.DialTelnet(flag.Arg(0), *username, *password, &jnpr.JnprIOProc{})
	if err != nil {
		panic(err)
	}

	defer s.Close()

	fmt.Printf("Server Capabilities: '%+v'\n", s.ServerCapabilities)
	fmt.Printf("Session Id: %d\n\n", s.SessionID)

	reply, err := s.Exec(netconf.MethodGetConfig("running"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply: %+v", reply)

}
