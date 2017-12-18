/*
 * Implementation of the client in Path ORAM by Stefanov et. al.
 */
package oram2pc

import (
	"fmt"
	"math"
)

/*
 * The client
 */
type Client struct {
	stash []Block
	pos map[int]int
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

		for j := 0; j < len(c.stash[i]); j++ {
			fmt.Printf(" [%b]", c.stash[i][j])
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
 *   B: Number of bytes in each block
 *   Z: Number of blocks in each bucket
 *   S: size of client's stash in blocks
 */
func Init_client(N int, B int, Z int, S int) *Client {
	c := &Client{stash: make([]Block, S, S), pos: make(map[int]int)}

	// initialize stash as all zeroes
	for i := range c.stash {
		c.stash[i] = make([]byte, B, B)
	}

	// initialize pos map as random values
	// create cryptographically secure shuffling of leaves
	L := int(math.Ceil(math.Log2(float64(N))))
	random_leaves := RandomPerm(1 << uint(L))
	fmt.Println(random_leaves)

	// assign each block with a unique random leaf
	for i := 0; i < N; i++ {
		// not all leaves will be used if N < 2^L but that's okay
		c.pos[i] = int(random_leaves[i])
	}

	return c
}

