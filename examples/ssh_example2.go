package main

import (
	"fmt"
	"log"

	"github.com/Juniper/go-netconf/netconf"
)

func main() {
	s, err := netconf.DialSSH("1.1.1.1", netconf.SSHConfigPassword("myuser", "mypass"))

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	fmt.Println(s.ServerCapabilities)
	fmt.Println(s.SessionID)

	reply, err := s.Exec(netconf.MethodGetConfig("running"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply: %+v", reply)
}
