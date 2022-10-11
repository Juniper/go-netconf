package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"time"

	netconf "github.com/nemith/netconf"
	ncssh "github.com/nemith/netconf/transport/ssh"
	"go.uber.org/goleak"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <server addr>\n", os.Args[0])
		os.Exit(1)
	}
	addr := os.Args[1]

	ctx := context.Background()

	usr, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user")
	}

	// ssh-agent(1) provides a UNIX socket at $SSH_AUTH_SOCK.
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		log.Fatalf("Failed to open SSH_AUTH_SOCK: %v", err)
	}

	agentClient := agent.NewClient(conn)
	config := &ssh.ClientConfig{
		User: usr.Username,
		Auth: []ssh.AuthMethod{
			// Use a callback rather than PublicKeys so we only consult the
			// agent once the remote server wants it.
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Add a connection timeout of 5 seconds.  You can also accomplish the same
	//  behavior with Timeout field of the ssh.ClientConfig.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	transport, err := ncssh.Dial(ctx, "tcp", addr, config)
	if err != nil {
		panic(err)
	}
	//defer transport.Close()

	session, err := netconf.Open(transport)
	if err != nil {
		panic(err)
	}
	//defer session.Close(context.Background())

	// timeout for the call itself.
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	reply, err := session.Do(ctx, &netconf.RPCMsg{Operation: `<get-configuration format="text"/>`})
	if err != nil {
		panic(err)
	}
	fmt.Println(reply.Data)

	deviceConfig, err := session.GetConfig(ctx, "running")
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}

	log.Printf("Config:\n%s\n", deviceConfig)

	if err := session.Close(context.Background()); err != nil {
		log.Print(err)
	}

	if err := goleak.Find(); err != nil {
		fmt.Println(err)
	}

}
