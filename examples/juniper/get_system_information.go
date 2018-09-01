// Go NETCONF Client - Juniper Example (show system information)
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/Juniper/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	host         = flag.String("host", "vmx1", "Hostname")
	username     = flag.String("username", "", "Username")
	key          = flag.String("key", os.Getenv("HOME")+"/.ssh/id_rsa", "SSH private key file")
	passphrase   = flag.String("passphrase", "", "SSH private key passphrase (cleartext)")
	nopassphrase = flag.Bool("nopassphrase", false, "SSH private key does not contain a passphrase")
	pubkey       = flag.Bool("pubkey", false, "Use SSH public key authentication")
	useAgent     = flag.Bool("agent", false, "Use SSH agent for public key authentication")
)

type SystemInformation struct {
	HardwareModel string `xml:"system-information>hardware-model"`
	OsName        string `xml:"system-information>os-name"`
	OsVersion     string `xml:"system-information>os-version"`
	SerialNumber  string `xml:"system-information>serial-number"`
	HostName      string `xml:"system-information>host-name"`
}

func sshPubKeyAgentConfig(user string) (*ssh.ClientConfig, error) {
	c, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agent.NewClient(c).Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func sshPubKeyFileConfig(user string, file string, passphrase string) (*ssh.ClientConfig, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	block, rest := pem.Decode(buf)
	if len(rest) > 0 {
		return nil, fmt.Errorf("pem: unable to decode file %s", file)
	}

	if x509.IsEncryptedPEMBlock(block) {
		b := block.Bytes
		b, err = x509.DecryptPEMBlock(block, []byte(passphrase))
		if err != nil {
			return nil, err
		}
		buf = pem.EncodeToMemory(&pem.Block{
			Type:  block.Type,
			Bytes: b,
		})
	}

	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil

}

func buildConfig() *ssh.ClientConfig {
	var config *ssh.ClientConfig
	var pass string
	if *pubkey == true {
		if *useAgent {
			var err error
			config, err = sshPubKeyAgentConfig(*username)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			if *nopassphrase {
				pass = "\n"
			} else {
				if *passphrase != "" {
					pass = *passphrase
				} else {
					var readpass []byte
					var err error
					fmt.Printf("Enter Passphrase for %s: ", *key)
					readpass, err = terminal.ReadPassword(int(syscall.Stdin))
					if err != nil {
						log.Fatal(err)
					}
					pass = string(readpass)
					fmt.Println()
				}
			}
			var err error
			config, err = netconf.SSHConfigPubKeyFile(*username, *key, pass)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		fmt.Printf("Enter Password: ")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println()

		config = &ssh.ClientConfig{
			User: *username,
			Auth: []ssh.AuthMethod{
				ssh.Password(string(password)),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}
	return config
}

func main() {
	flag.Parse()

	if *username == "" {
		fmt.Printf("Enter a valid username: ")
		r := bufio.NewScanner(os.Stdin)
		r.Scan()
		*username = r.Text()
	}

	sshConfig := buildConfig()
	s, err := netconf.DialSSH(*host, sshConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	reply, err := s.Exec(netconf.RawMethod("<get-system-information/>"))
	if err != nil {
		panic(err)
	}
	var q SystemInformation

	xml.Unmarshal([]byte(reply.RawReply), &q)
	fmt.Printf("hostname: %s\n", q.HostName)
	fmt.Printf("model: %s\n", q.HardwareModel)
	fmt.Printf("version: %s\n", q.OsVersion)
}
