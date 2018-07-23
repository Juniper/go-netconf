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
	"testing"
)

func BenchmarkLog(b *testing.B) {
	log.SetFlags(log.LstdFlags)
	file, err := ioutil.TempFile("", "benchmark-log")
	if err != nil {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(file)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Printf("Benchmark %d", i)
	}
}

func BenchmarkNopLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Logger.Println("Benchmark %d", i)
	}
}
