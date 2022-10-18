package inttest

import (
	"context"
	"os"
	"testing"

	"github.com/nemith/netconf"
	ncssh "github.com/nemith/netconf/transport/ssh"
	"golang.org/x/crypto/ssh"
)

func sshAuth(t *testing.T) ssh.AuthMethod {
	switch {
	case os.Getenv("NETCONF_DUT_SSHPASS") != "":
		return ssh.Password(os.Getenv("NETCONF_DUT_SSHPASS"))
	case os.Getenv("NETCONF_DUT_SSHKEYFILE") != "":
		keyFile := os.Getenv("NETCONF_DUT_SSHKEYFILE")
		key, err := os.ReadFile(keyFile)
		if err != nil {
			t.Fatalf("couldn't open ssh private key %q: %v", keyFile, err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			t.Fatalf("couldn't parse private key %q: %v", keyFile, err)
		}

		return ssh.PublicKeys(signer)
	}
	t.Fatal("NETCONF_DUT_SSHADDR tests require NETCONF_DUT_SSHPASS or NETCONF_DUT_SSHKEYFILE")
	return nil
}

func TestSSHConnect(t *testing.T) {
	addr := os.Getenv("NETCONF_DUT_SSHADDR")
	if addr == "" {
		t.Skip("NETCONF_DUT_SSHADDR not set, skipping test")
	}

	user := os.Getenv("NETCONF_DUT_SSHUSER")
	if user == "" {
		t.Fatal("NETCONF_DUT_SSHADDR set but NETCONF_DUT_SSHUSER is not set")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{sshAuth(t)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	t.Logf("connecting to %s", addr)

	ctx := context.Background()
	tr, err := ncssh.Dial(ctx, "tcp", addr, config)
	if err != nil {
		t.Fatalf("failed to connect to dut: %v", err)
	}

	// capture the framed communication
	inCap := newLogWriter("<--", t)
	outCap := newLogWriter("-->", t)

	defer func() {
		inCap.Close()
		outCap.Close()
	}()

	tr.DebugCapture(inCap, outCap)

	session, err := netconf.Open(tr)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if session.SessionID() <= 0 {
		t.Fatalf("invalid session id: %d", session.SessionID())
	}

	if len(session.ServerCapabilities()) <= 0 {
		t.Fatalf("invalid server capabilities for session")
	}

	// XXX: GetConfig

	if err := session.Close(ctx); err != nil {
		t.Fatalf("failed to close session: %v", err)
	}
}
