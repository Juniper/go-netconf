package jnpr

import (
	"fmt"
	"regexp"

	"github.com/Juniper/go-netconf/netconf"
)

const (
	netconfCMDCLI   = "junoscript netconf need-trailer"
	netconfCMDShell = "exec xml-mode netconf need-trailer"
)

const (
	// CLIModeShell tells the system what mode to open the shell in
	CLIModeShell = iota
	// CLIModeCLI sets (no) mode for the CLI
	CLIModeCLI
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
		j.cliMode = CLIModeCLI
	case "%":
		j.cliMode = CLIModeShell
	default:
		return fmt.Errorf("Cannot determine prompt '%s'", prompt[0])
	}
	return nil
}

func (j *JnprIOProc) StartNetconf(t *netconf.TransportTelnet) error {
	switch j.cliMode {
	case CLIModeShell:
		t.Writeln([]byte(netconfCMDShell))
		return nil
	case CLIModeCLI:
		t.Writeln([]byte(netconfCMDCLI))
		return nil
	}
	return fmt.Errorf("Cannot start netconf: Unknown CLI mode '%d'", j.cliMode)
}
