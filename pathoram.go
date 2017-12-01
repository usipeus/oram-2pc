/*
 * Implementation of Path ORAM by Stefanov et. al.
*/
package oram2pc

import "bytes"

type Block struct {
	mem []byte
}

/*
 * The client
*/
type Client struct {
	stash []Block
}

/*
 * Pretty print function for a client
 *
 * Prints the stash and various runtime statistics
*/
func (c *Client) pprint() {

}

