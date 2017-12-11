/*
 * Implementation of the server in Path ORAM by Stefanov et. al.
 */
package oram2pc

import (
	// "fmt"
	// "math/rand"
	"os"
)

type Block struct {
	mem []byte
}

/*
 * Each bucket contains Z blocks; it suffices that Z is small e.g. 4
 */
type Bucket struct {
	blk []Block
}

/*
 * The server
 */
type Server struct {
	dir string
}

/*
 * Initialize a Server
 *
 * Returns a new Server with params:
 *   B: Capacity of each block in bytes
 *   Z: Capacity of each bucket in blocks
 */
func Init_server(B uint, Z uint) *Server {
	s := &Server{}

	return s
}

