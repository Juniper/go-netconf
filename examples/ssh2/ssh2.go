package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"

	"github.com/Juniper/go-netconf/netconf"
)

func main() {
	sshConfig := &ssh.ClientConfig{
		User: "myuser",
		Auth: []ssh.AuthMethod{ssh.Password("mypass")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	s, err := netconf.DialSSH("1.1.1.1", sshConfig)

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
