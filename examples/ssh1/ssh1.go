// Go NETCONF Client - Example
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Juniper/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
)

func main() {
	sshConfig := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password("xxx")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	s, err := netconf.DialSSH("172.16.240.189", sshConfig)

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	// enable RPC logging
	netconf.Logger.SetOutput(os.Stderr)

	fmt.Println(s.ServerCapabilities)
	fmt.Println(s.SessionID)

	// Sends raw XML
	reply, err := s.Exec(netconf.RawMethod("<get-chassis-inventory/>"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply: %+v", reply)
}
