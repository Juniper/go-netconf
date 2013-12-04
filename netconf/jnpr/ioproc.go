package jnpr

import (
	"fmt"
	"github.com/Juniper/go-netconf/netconf"
	"regexp"
)

const (
	NETCONF_CMD_CLI   = "junoscript netconf need-trailer"
	NETCONF_CMD_SHELL = "exec xml-mode netconf need-trailer"
)

const (
	CLI_MODE_SHELL = iota
	CLI_MODE_CLI
)

type JnprIOProc struct {
	cliMode int
}

var promptRE = regexp.MustCompile(`([>%])\s+`)

func (j *JnprIOProc) Login(t *netconf.TransportTelnet, username string, password string) error {
	t.WaitForString("ogin:")
	t.Writeln([]byte(username))

	t.WaitForString("assword:")
	t.Writeln([]byte(password))

	_, prompt, err := t.WaitForRegexp(promptRE)
	if err != nil {
		return err
	}

	switch string(prompt[0]) {
	case ">":
		j.cliMode = CLI_MODE_CLI
	case "%":
		j.cliMode = CLI_MODE_SHELL
	default:
		return fmt.Errorf("Cannot determine prompt '%s'", prompt[0])
	}
	return nil
}

func (j *JnprIOProc) StartNetconf(t *netconf.TransportTelnet) error {
	switch j.cliMode {
	case CLI_MODE_SHELL:
		t.Writeln([]byte(NETCONF_CMD_SHELL))
		return nil
	case CLI_MODE_CLI:
		t.Writeln([]byte(NETCONF_CMD_CLI))
		return nil
	}
	return fmt.Errorf("Cannot start netconf: Unknown CLI mode '%s'", j.cliMode)
}
