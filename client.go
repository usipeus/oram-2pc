/*
 * Implementation of the client in Path ORAM by Stefanov et. al.
 */
package oram2pc

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

/*
 * The client
 */
type Client struct {
	N          int
	L          int
	B          int
	Z          int
	stash      map[string][]Block
	stash_free map[string]int
	pos        map[int]int
	keys       map[string][]byte
	servers    map[string]*Server
}

/*
 * Pretty print function for a client
 *
 * Prints the stash and various runtime statistics
 */
func (c *Client) Pp(name string) {
	// print blocks
	stash := c.stash[name]
	for i := 0; i < len(stash); i++ {
		fmt.Printf("{")

		for j := 0; j < len(stash[i]); j++ {
			fmt.Printf(" [%b]", stash[i][j])
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
 *   L: height of the tree
 *   B: Number of bytes in each block (fixed to 32 bytes)
 *   Z: Number of blocks in each bucket
 *   S: size of client's stash in blocks
 *   key: 16-byte key to encrypt blocks with
 */
func InitClient(N int, Z int) *Client {
	c := &Client{N: N, B: 32, Z: Z, pos: make(map[int]int)}

	// initialize pos map as random values
	// create cryptographically secure shuffling of leaves
	c.L = int(math.Ceil(math.Log2(float64(N))))
	random_leaves := random_perm(1 << uint(c.L))

	// assign each block with a unique random leaf
	for i := 0; i < N; i++ {
		// not all leaves will be used if N < 2^L but that's okay
		c.pos[i] = int(random_leaves[i])
	}

	// init stash map
	c.stash = make(map[string][]Block)
	c.stash_free = make(map[string]int)

	// initialize empty server map and keys map
	c.servers = make(map[string]*Server)
	c.keys = make(map[string][]byte)

	return c
}

func (c *Client) ServerInfo(name string) string {
	s, prs := c.servers[name]
	if prs == false {
		return ""
	}

	namestr := "Server: " + name
	nstr := "\tN: " + strconv.Itoa(s.N)
	zstr := "\tZ: " + strconv.Itoa(s.Z)
	dirstr := "\tdir: " + s.dir

	return strings.Join([]string{namestr, nstr, zstr, dirstr}, "\n")
}

func (c *Client) AddServer(name string, N int, Z int, fsize int) error {
	_, prs := c.servers[name]
	if prs == true {
		return errors.New("A server already exists with that name!")
	}
	// add new server
	c.servers[name] = init_server(N, Z, fsize)

	// generate random key for that server
	key := make([]byte, 16)
	rand.Read(key)
	c.keys[name] = key

	// init stash
	S := c.N * c.L
	cur_stash := make([]Block, S)

	// initialize stash as dummy blocks
	for i := range cur_stash {
		cur_stash[i] = enc_dummy_block(key)
	}

	c.stash[name] = cur_stash
	c.stash_free[name] = S

	// initialize serverside storage as all dummy blocks
	err := c.init_server_storage(name, key)

	return err
}

func (c *Client) init_server_storage(name string, key []byte) error {
	s, prs := c.servers[name]
	if prs == false {
		return errors.New("No server found by that name!")
	}

	s.create_tree()

	// encrypt c.Z dummy blocks to get a bucket, and write to every node in tree
	for i := 0; i <= s.L; i++ {
		for j := 0; j < (1 << uint(i)); j++ {
			bucket := make_bucket(nil, c.Z, key)
			s.write_node(bucket, i, j)
		}
	}

	return nil
}

func (c *Client) Access(name string, op bool, a int, data uint64) (uint64, error) {
	var ret uint64 = 0

	// get server
	s, prs := c.servers[name]
	if prs == false {
		return ret, errors.New("Could not find server by that name!")
	}

	// get encryption key for this server
	key, prs := c.keys[name]
	if prs == false {
		return ret, errors.New("Could not find server by that name!")
	}

	// get position from posmap
	x, prs := c.pos[a]
	if prs == false {
		return ret, errors.New("Tried to look up invalid block number in pos!")
	}

	// map block a to new random leaf
	num_leaves := 1 << uint(c.L)
	new_leaf := gen_int(num_leaves)
	c.pos[a] = new_leaf

	// read path containing block a (i.e. the path to leaf x)
	buckets, err := s.get_path_buckets(x)
	if err != nil {
		return ret, err
	}

	// find which bucket contains a (the block we want to update)
	bucket_idx, ret := find_block(buckets, a, key)
	fmt.Printf("finding which bucket contains a: found %i\n", bucket_idx)
	fmt.Printf("read value %u\n", ret)

	// if op is 1 (write), update bucket that contains the data to write
	if op == true {
		new_block := block_encode(a, data)
		new_bucket := make_bucket([]Block{new_block}, s.Z, key)

		if bucket_idx != -1 && ret == 0 {
			buckets[bucket_idx] = new_bucket
		} else {
			// update any block since they're all dummy blocks
			buckets[0] = new_bucket
		}
	}

	// find intersections between old and new path
	old_path, err := s.get_path(x)
	if err != nil {
		return ret, err
	}
	fmt.Println("old path:", old_path)

	new_path, err := s.get_path(new_leaf)
	if err != nil {
		return ret, err
	}
	fmt.Println("new path:", new_path)

	num_inters := 0
	for i := range old_path {
		if old_path[i] == new_path[i] {
			num_inters += 1
		}
	}

	// fill list of intersections greedily (starting from the leaf)
	inters := make([]int, num_inters)
	for i := 0; i < num_inters; i++ {
		inters[i] = old_path[num_inters-1-i]
	}
	fmt.Println("Found intersection:", inters)

	// write back path

	// write back nondummy blocks first
	blk_to_write := find_nondummy(buckets, key)
	fmt.Println("Writing back nondummy blocks:", blk_to_write)
	cur_stash := c.stash[name]

	// if there aren't enough blocks to write back, then write stash blocks
	for c.stash_free[name] < s.N*s.L {
		if len(blk_to_write) >= len(inters) {
			break
		}

		blk_to_write = append(blk_to_write, cur_stash[:1]...)
		cur_stash = append(cur_stash[1:])
		c.stash_free[name] += 1
		fmt.Printf("Appended a stash block...\n")
	}

	// if out of stash blocks, write dummy blocks
	if len(blk_to_write) < len(inters) {
		dummy := make([]Block, len(inters)-len(blk_to_write))
		for i := range dummy {
			dummy[i] = dummy_block()
		}

		blk_to_write = append(blk_to_write, dummy...)
		fmt.Printf("Appending %i dummy blocks...\n", len(dummy))
	}

	// write blocks
	for i := range inters {
		cur_l := len(inters) - 1 - i
		bucket := make_bucket([]Block{blk_to_write[i]}, s.Z, key)
		s.write_node(bucket, cur_l, inters[i])
	}

	// if can't write anymore blocks, copy to the stash
	if len(inters) < len(blk_to_write) {
		extra_blks := blk_to_write[len(inters):]
		cur_stash = append(cur_stash, extra_blks...)
		c.stash_free[name] -= len(extra_blks)

		fmt.Printf("Adding %i blocks to stash...\n", len(extra_blks))
	}

	c.stash[name] = cur_stash

	return ret, nil
}
