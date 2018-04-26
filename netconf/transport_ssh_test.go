// Go NETCONF Client
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netconf

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestSSHConfigPassword(t *testing.T) {
	user := "test"
	password := "testPass"

	res := SSHConfigPassword(user, password)

	// SSHConfigPassword needs to return ssh.ClientConfig
	scpType := reflect.TypeOf(res).String()
	scpDesiredType := "*ssh.ClientConfig"
	if scpType != scpDesiredType {
		t.Errorf("got type %s, expected %s", scpType, scpDesiredType)
	}

	// User is set and correct
	if res.User != user {
		t.Errorf("got user %s, expected %s", res.User, user)
	}

	// Auth method is password
	authMethodType := reflect.TypeOf(res.Auth[0]).String()
	authMethodDesiredType := "ssh.passwordCallback"
	if authMethodType != authMethodDesiredType {
		t.Errorf("got type %s, expected %s", authMethodType, authMethodDesiredType)
	}

	// Ignre host key
	hostKeyMethod := runtime.FuncForPC(reflect.ValueOf(res.HostKeyCallback).Pointer()).Name()
	if !strings.Contains(hostKeyMethod, "InsecureIgnoreHostKey") {
		t.Errorf("host key method of %s does not contain expected InsecureIgnoreHostKey", hostKeyMethod)
	}
}
