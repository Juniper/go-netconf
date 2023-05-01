//go:build inttest
// +build inttest

package inttest

import (
	"context"
	"encoding/xml"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/nemith/netconf"
	ncssh "github.com/nemith/netconf/transport/ssh"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func onlyFlavor(t *testing.T, flavors ...string) {
	t.Helper()
	for _, flavor := range flavors {
		if os.Getenv("NETCONF_DUT_FLAVOR") == flavor {
			return
		}
	}
	t.Skipf("test only for flavors '%s'.  Skipping", strings.Join(flavors, ","))
}

func sshAuth(t *testing.T) ssh.AuthMethod {
	t.Helper()

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

func setupSSH(t *testing.T) *netconf.Session {
	t.Helper()

	host := os.Getenv("NETCONF_DUT_SSHHOST")
	if host == "" {
		t.Skip("NETCONF_DUT_SSHHOST not set, skipping test")
	}

	port := os.Getenv("NETCONF_DUT_SSHPORT")
	if port == "" {
		port = "830"
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

	addr := net.JoinHostPort(host, port)
	t.Logf("connecting to %s", addr)

	ctx := context.Background()
	tr, err := ncssh.Dial(ctx, "tcp", addr, config)
	if err != nil {
		t.Fatalf("failed to connect to dut: %v", err)
	}

	// capture the framed communication
	inCap := newLogWriter("<<<", t)
	outCap := newLogWriter(">>>", t)

	tr.DebugCapture(inCap, outCap)

	session, err := netconf.Open(tr)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return session
}

func TestSSHGetConfig(t *testing.T) {
	session := setupSSH(t)

	if session.SessionID() <= 0 {
		t.Fatalf("invalid session id: %d", session.SessionID())
	}

	if len(session.ServerCapabilities()) == 0 {
		t.Fatalf("invalid server capabilities for session")
	}

	ctx := context.Background()
	config, err := session.GetConfig(ctx, "running")
	if err != nil {
		t.Errorf("failed to call get-config: %v", err)
	}
	t.Logf("configuration: %s", config)

	// XXX: GetConfig
	if err := session.Close(ctx); err != nil {
		t.Fatalf("failed to close session: %v", err)
	}
}

func TestBadGetConfig(t *testing.T) {
	session := setupSSH(t)

	ctx := context.Background()
	cfg, err := session.GetConfig(ctx, "non-exist")
	assert.Nil(t, cfg)
	var rpcErr netconf.RPCError
	assert.ErrorAs(t, err, &rpcErr)
}

func TestJunosCommand(t *testing.T) {
	onlyFlavor(t, "junos")
	session := setupSSH(t)

	cmd := struct {
		XMLName xml.Name `xml:"command"`
		Command string   `xml:",innerxml"`
	}{
		Command: "show version",
	}

	ctx := context.Background()
	err := session.Call(ctx, &cmd, nil)
	assert.NoError(t, err)
}
