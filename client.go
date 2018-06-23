package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"./netconf"

	"golang.org/x/crypto/ssh"
)

const versionstring string = "0.2"

func main() {

	var version = flag.Bool("version", false, "Show version for code")
	var targethost = flag.String("targethost", "", "Target IPv4 or FQDN hostname of a NETCONF node")
	/*var transport = */ flag.String("transport", "ssh", "Transport mode (NOT IMPLEMENTED)")
	var envelope = flag.String("envelope", "", "XML Envelope")
	var envelopefile = flag.String("envelopefile", "", "XML envelope file path")
	var username = flag.String("username", "", "Username")
	var password = flag.String("password", "", "Password")
	var port = flag.Int("port", 830, "Port for accessing NETCONF")
	/*var sshkey = */ flag.String("sshkey", "", "SSHKey for accessing node (NOT IMPLEMENTED)")
	flag.Parse()

	// If version has been requested
	if *version {
		fmt.Printf("Version is: %s", versionstring)
		os.Exit(0)
	}

	// TODO(davidjohngee)
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

	// netconfcall content var
	netconfcall := ""
	// If the flag 'envelopefile' is set, then let's read the content into memory instead of using envelope
	if *envelopefile != "" {
		// Read content from disk
		bytes, err := ioutil.ReadFile(*envelopefile)
		netconfcall = string(bytes)
		if err != nil {
			log.Panic(err)
		}
		// Else, consume content of envelope
	} else {
		netconfcall = *envelope
	}

	reply, err := s.Exec(netconf.RawMethod(netconfcall))
	if err != nil {
		panic(err)
	}
	fmt.Print(reply.Data)
}
