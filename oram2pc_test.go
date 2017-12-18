/*
 * Test functions for the oram2pc library
 */
package oram2pc

import (
	"fmt"
	"testing"
)

func Test_pathoram(t *testing.T) {
	c := Init_client(6, 256 / 8, 4, 4)
	c.Pp()

	s := Init_server(6, 256 / 8, 4, 4 * 1024)
	fmt.Println(s.dir)

	s.get_path(4)
}


