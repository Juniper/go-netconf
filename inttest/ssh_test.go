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
	"github.com/stretchr/testify/require"
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
		require.NoErrorf(t, err, "couldn't open ssh private key %q", keyFile)

		signer, err := ssh.ParsePrivateKey(key)
		require.NoError(t, err)
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
	require.NoErrorf(t, err, "failed to connect to dut %q", addr)

	// capture the framed communication
	inCap := newLogWriter("<<<", t)
	outCap := newLogWriter(">>>", t)

	tr.DebugCapture(inCap, outCap)

	session, err := netconf.Open(tr)
	require.NoError(t, err, "failed to create netconf session")
	return session
}

func TestSSHOpen(t *testing.T) {
	session := setupSSH(t)
	assert.NotZero(t, session.SessionID())
	assert.NotEmpty(t, session.ServerCapabilities())
	err := session.Close(context.Background())
	assert.NoError(t, err)
}

func TestSSHGetConfig(t *testing.T) {
	session := setupSSH(t)

	ctx := context.Background()
	config, err := session.GetConfig(ctx, "running")
	assert.NoError(t, err)
	t.Logf("configuration: %s", config)

	err = session.Close(ctx)
	assert.NoError(t, err)
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
