/*
 * With this example the program fetches all of the yang models from netconf, then
 * it parses the yang models with libyang and with the libyang's API it transforms
 * an XPATH into a XML for Netconf.
 *
 * Example:
 *
 * "/ietf-interfaces:interfaces/interface[name='eth1']/ietf-ip:ipv4"
 *
 * <interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">
 *     <interface>
 *         <name>eth1</name>
 *         <ipv4 xmlns="urn:ietf:params:xml:ns:yang:ietf-ip"/>
 *     </interface>
 * </interfaces>
 */

package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"unsafe"

	"github.com/sartura/go-netconf/netconf"
)

/*
#cgo LDFLAGS: -lyang
#cgo LDFLAGS: -lpcre
#include <libyang/libyang.h>
#include <libyang/tree_data.h>

#include <stdlib.h>
#include "helper.h"
*/
import "C"

type Data struct {
	XMLName xml.Name `xml:"data"`
	Schema  []Schema `xml:"netconf-state>schemas>schema"`
}

type Schema struct {
	XMLName    xml.Name `xml:"schema"`
	Identifier string   `xml:"identifier"`
	Version    string   `xml:"version"`
	Format     string   `xml:"format"`
	Namespace  string   `xml:"namespace"`
	Location   string   `xml:"location"`
}

func getRemoteContext(s *netconf.Session) *C.struct_ly_ctx {
	var err error
	ctx := C.ly_ctx_new(nil)

	getSchemas := `
	<get>
	<filter type="subtree">
	<netconf-state xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring">
	<schemas/>
	</netconf-state>
	</filter>
	</get>
	`
	// Sends raw XML
	reply, err := s.Exec(netconf.RawMethod(getSchemas))
	if err != nil {
		panic(err)
	}

	var data Data
	err = xml.Unmarshal([]byte(reply.Data), &data)
	if err != nil {
		panic(err)
	}

	getSchema := `
	<get-schema xmlns="urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring"><identifier>%s</identifier><version>%s</version><format>%s</format></get-schema>
	`
	for i := range data.Schema {
		if data.Schema[i].Format == "yang" {
			schema := data.Schema[i]
			request := fmt.Sprintf(getSchema, schema.Identifier, schema.Version, schema.Format)
			reply, err := s.Exec(netconf.RawMethod(request))
			if err != nil {
				fmt.Printf("init data ERROR: %s\n", err)
			}
			var yang string
			err = xml.Unmarshal([]byte(reply.Data), &yang)
			if err != nil {
				panic(err)
			}
			_ = C.lys_parse_mem(ctx, C.CString(yang), C.LYS_IN_YANG)
		}
	}
	return ctx
}

//export GoErrorCallback
func GoErrorCallback(level C.LY_LOG_LEVEL, msg *C.char, path *C.char) {
	log.Printf("libyang error: %s\n", C.GoString(msg))
	return
}

func main() {

	// set error callback for libyang
	C.ly_set_log_clb((C.clb)(unsafe.Pointer(C.CErrorCallback)), 0)

	auth := netconf.SSHConfigPassword("root", "root")
	s, err := netconf.DialSSH("0.0.0.0:830", auth)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// create new libyang context with the remote yang files
	ctx := getRemoteContext(s)
	// free context at exit
	defer C.ly_ctx_destroy(ctx, nil)

	node := C.lyd_new_path(nil, ctx, C.CString("/ietf-interfaces:interfaces/interface[name='eth1']/ietf-ip:ipv4"), nil, C.LYD_ANYDATA_XML, 0)
	if node == nil {
		fmt.Printf("Error")
		return
	}
	defer C.lyd_free_withsiblings(node)

	var str *C.char
	C.lyd_print_mem(&str, node, C.LYD_XML, 0)
	defer C.free(unsafe.Pointer(str))

	fmt.Printf("Generated xml from xpath:\n%s\n", C.GoString(str))
}
