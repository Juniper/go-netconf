package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"os"
	"time"

	netconf "github.com/nemith/netconf"
	nctls "github.com/nemith/netconf/transport/tls"
)

var (
	//go:embed certs/ca.crt
	caCert []byte

	//go:embed certs/client.crt
	clientCert []byte

	//go:embed certs/client.key
	clientKey []byte
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <server addr>\n", os.Args[0])
		os.Exit(1)
	}
	addr := os.Args[1]

	ctx := context.Background()

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		panic(err)
	}

	// Add a connection timeout of 5 seconds.  You can also accomplish the same
	//  behavior with Timeout field of the ssh.ClientConfig.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	config := tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
	}

	transport, err := nctls.Dial(ctx, "tcp", addr, &config)
	if err != nil {
		panic(err)
	}
	defer transport.Close()

	session, err := netconf.Open(transport)
	if err != nil {
		panic(err)
	}
	defer session.Close(context.Background())

	// timeout for the call itself.
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	reply, err := session.Do(ctx, &netconf.RPCMsg{Operation: &netconf.GetConfig{Source: "running"}})
	/* reply, err := session.Do(ctx, "<get-config><source><running><running/></source></get-config>") */
	if err != nil {
		panic(err)
	}
	fmt.Println(reply.Data)
}
