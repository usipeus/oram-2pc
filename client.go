/*
 * Implementation of the client in Path ORAM by Stefanov et. al.
 */
package oram2pc

import (
	"fmt"
)

/*
 * The client
 */
type Client struct {
	stash []Block
	pos map[uint]uint
}

/*
 * Pretty print function for a client
 *
 * Prints the stash and various runtime statistics
 */
func (c *Client) Pp() {
	// print blocks
	for i := 0; i < len(c.stash); i++ {
		fmt.Printf("{")

		for j := 0; j < len(c.stash[i].mem); j++ {
			fmt.Printf(" [%b]", c.stash[i].mem[j])
		}

		fmt.Printf(" }\n")
	}

	fmt.Println(c.pos)
}

/*
 * Initialize a Client
 *
 * Returns a new Client with params:
 *   N: number of blocks outsourced to server
 *   L: height of binary tree
 *   S: size of client's stash in blocks
 *   B: Number of bytes in each block
 */
func Init_client(N uint, L uint, S uint, B uint) *Client {
	c := &Client{stash: make([]Block, S, S), pos: make(map[uint]uint)}

	// initialize stash as all zeroes
	for i := range c.stash {
		c.stash[i].mem = make([]byte, B, B)
	}

	// initialize pos map as random values
	// create cryptographically secure shuffling of leaves
	random_leaves := RandomPerm(1 << L)
	fmt.Println(random_leaves)

	// assign each block with a unique random leaf
	var i uint
	for i = 0; i < N; i++ {
		// not all leaves will be used if N < 2^L but that's okay
		c.pos[i] = uint(random_leaves[i]);
	}

	return c
}

