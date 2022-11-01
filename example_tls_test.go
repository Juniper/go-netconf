package netconf_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nemith/netconf"
	nctls "github.com/nemith/netconf/transport/tls"
)

const tlsAddr = "myrouter.example.com:6513"

func Example_tls() {
	caCert, err := os.ReadFile("ca.crt")
	if err != nil {
		log.Fatalf("failed to load ca cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCert, err := os.ReadFile("client.crt")
	if err != nil {
		log.Fatalf("failed to load client cert: %v", err)
	}

	clientKey, err := os.ReadFile("client.key")
	if err != nil {
		log.Fatalf("failed to load client key: %v", err)
	}

	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		panic(err)
	}

	// tls transport configuration
	config := tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
	}

	// Add a connection establish timeout of 5 seconds.  You can also accomplish
	// the same behavior with Timeout field of the ssh.ClientConfig.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	transport, err := nctls.Dial(ctx, "tcp", tlsAddr, &config)
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

	cfg, err := session.GetConfig(ctx, "running")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Config: %s\n", cfg)

	if err := session.Close(context.Background()); err != nil {
		log.Print(err)
	}
}
