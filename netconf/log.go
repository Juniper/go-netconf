// Go NETCONF Client
//
// Copyright (c) 2013-2018, Juniper Networks, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netconf

import (
	"io/ioutil"
	"log"
	"os"
)

func init() {
	// Discard is an io.Writer on which all Write calls succeed without doing anything.
	Logger.SetOutput(ioutil.Discard)
}

// NopLogger is a wrapper for the standard logger
type NopLogger struct {
	*log.Logger
}

// Logger provides request and response payload logging for the Netconf Exec
var Logger = &NopLogger{
	log.New(os.Stderr, "", log.LstdFlags),
}
