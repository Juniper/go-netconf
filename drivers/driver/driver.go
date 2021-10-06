// Copyright (c) 2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netconf

import rpc "github.com/davedotdev/go-netconf/rpc"

// Driver interface for building drivers that are self-contained from a user's perspective.
type Driver interface {
	Lock(ds string) (*rpc.RPCReply, error)
	Unlock(ds string) (*rpc.RPCReply, error)

	Close() error
	Dial() error
	DialTimeout() error
	SendRaw(rawxml string) (*rpc.RPCReply, error)
	GetConfig() (*rpc.RPCReply, error)
}

// New is an interface that checks compliancy
func New(d Driver) Driver {
	return d
}
