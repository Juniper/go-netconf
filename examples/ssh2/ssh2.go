// Go NETCONF Client - Example
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/inspiregates/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
)

func main() {
	sshConfig := &ssh.ClientConfig{
		User:            "myuser",
		Auth:            []ssh.AuthMethod{ssh.Password("mypass")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	s, err := netconf.DialSSH("1.1.1.1", sshConfig)

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	fmt.Println(s.ServerCapabilities)
	fmt.Println(s.SessionID)

	reply, err := s.Exec(netconf.MethodGetConfig("running"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply: %+v", reply)
}
