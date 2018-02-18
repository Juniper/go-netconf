package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/ssh"

	"github.com/Juniper/go-netconf/netconf"
)

func main() {
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.Password("xxx")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	s, err := netconf.DialSSH("172.16.240.189", sshConfig)

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	fmt.Println(s.ServerCapabilities)
	fmt.Println(s.SessionID)

	// Sends raw XML
	reply, err := s.Exec(netconf.RawMethod("<get-chassis-inventory/>"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply: %+v", reply)
}
