package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"

	"github.com/Juniper/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func getSessionID(addr string, sshConfig *ssh.ClientConfig) (int, error) {
	bastionClient, err := ssh.Dial("tcp", "bastion-host.example.net:22", sshConfig)
	if err != nil {
		return 0, fmt.Errorf("error dialing bastion host: %v", err)
	}
	defer bastionClient.Close()

	tunConn, err := bastionClient.Dial("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer tunConn.Close()

	sess, err := netconf.NewSSHSession(tunConn, addr, sshConfig)
	if err != nil {
		return 0, err
	}
	defer sess.Close()

	id := sess.SessionID
	return id, nil
}

func main() {
	/*
		Example of connecting via an SSH bastion host using SSH agent authentication
		and known_hosts public key checking.
	*/
	u, err := user.Current()
	if err != nil {
		log.Fatalf("error retrieving current user: %v\n", err)
	}

	kHosts, err := knownhosts.New(u.HomeDir + "/.ssh/known_hosts")
	if err != nil {
		log.Fatalf("error reading known hosts file: %v\n", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            "myuser",
		Auth:            []ssh.AuthMethod{sshAgent()},
		HostKeyCallback: kHosts,
	}

	id, err := getSessionID("netconf-device.corp.local:22", sshConfig)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}
