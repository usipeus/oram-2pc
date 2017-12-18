/*
 * Test functions for the oram2pc library
 */
package oram2pc

import (
	"fmt"
	"testing"
)

func Test_pathoram(t *testing.T) {
	// currently use 256-bit blocks to store encrypted 128-bit values
	c := Init_client(6, 4, 4)
	c.Pp()

	s := Init_server(6, 4, 4 * 1024)
	fmt.Println(s.dir)

	s.get_path(4)
	aoeu := enc_dummy_block([]byte("key lul"))
	b := enc_block(0x1122334455667788, []byte("key lul"))
	fmt.Println(b)
	d, dummy := dec_block(b, []byte("key lul"))
	if dummy == false {
		fmt.Printf("%x\n", d)
	}

	e, dummy := dec_block(aoeu, []byte("key lul"))
	if dummy == true {
		fmt.Println("found dummy block!", e)
	}
}
