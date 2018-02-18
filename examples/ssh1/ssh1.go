package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/ssh"

	"github.com/Juniper/go-netconf/netconf"
)

func main() {
	sshConfig := &ssh.ClientConfig{
		Config: ssh.Config{
			Ciphers: []string{"aes128-cbc", "hmac-sha1"},
		},
		User: "myuser",
		Auth: []ssh.AuthMethod{ssh.Password("mypass")},
	}

	s, err := netconf.DialSSH("1.1.1.1", sshConfig)

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
