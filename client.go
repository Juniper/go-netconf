package main

import (
	"flag"
	"fmt"
	"log"

	"./netconf"

	"golang.org/x/crypto/ssh"
)

func main() {

	var targethost = flag.String("targethost", "", "Target IPv4 or FQDN hostname of a NETCONF node")
	/*var transport = */ flag.String("transport", "ssh", "Transport mode (NOT IMPLEMENTED)")
	var envelope = flag.String("envelope", "", "XML Envelope")
	var username = flag.String("username", "", "Username")
	var password = flag.String("password", "", "Password")
	var port = flag.Int("port", 830, "Port for accessing NETCONF")
	/*var sshkey = */ flag.String("sshkey", "", "SSHKey for accessing node (NOT IMPLEMENTED)")
	flag.Parse()

	// TODO
	// Make some decisions over auth mechanism.
	// 1. If username & password is present, ignore SSH key
	// 2. If SSH key is present, ignore username and password
	// Current, we ignore sshkey totally and pass in username and password

	sshConfig := &ssh.ClientConfig{
		User:            *username,
		Auth:            []ssh.AuthMethod{ssh.Password(*password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	s, err := netconf.DialSSH(*targethost, sshConfig, *port)

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod(*envelope))
	if err != nil {
		panic(err)
	}
	fmt.Print(reply.Data)
}
