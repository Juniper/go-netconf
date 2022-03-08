// Go NETCONF Client - Example
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Juniper/go-netconf/netconf"
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

	var queryBuilder strings.Builder

	queryBuilder.WriteString(
		"<configuration><interfaces><interface><name>INTERFACE-NAME</name><unit><name>0</name><description>MANAGEMENT</description></unit></interface></interfaces></configuration>")

	write := netconf.MethodEditConfig("candidate", "stop-on-error", queryBuilder.String())

	reply, err := s.Exec(write)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply: %+v", reply)
}
