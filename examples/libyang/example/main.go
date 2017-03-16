/*
 * A simple example on how to include libyang library into Go
 */

package main

import (
	"fmt"
)

/*
#cgo LDFLAGS: -lyang
#cgo LDFLAGS: -lpcre
#include <libyang/libyang.h>
*/
import "C"

func main() {
	// create libyang context
	ctx := C.ly_ctx_new(C.CString("./"))

	if ctx != nil {
		fmt.Printf("Libyang context successfully created.\n")
	} else {
		fmt.Printf("Failed to load libyang.\n")
	}
}
