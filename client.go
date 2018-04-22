/*
 * Implementation of the client in Path ORAM by Stefanov et. al.
 */
package oram2pc

import (
	"crypto/rand"
	"errors"
	// "fmt"
	"math"
	// "net"
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
	pos        map[int]int
	keys       map[string][]byte
	servers    map[string]*Server
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
	c.stash[name] = make([]Block, 0, S)

	// initialize serverside storage as all dummy blocks
	err := c.init_server_storage(name, key)

	return err
}

func (c *Client) RemoveServer(name string) error {
	s, prs := c.servers[name]
	if prs == true {
		err := s.remove_tree()
		return err
	}

	return errors.New("No server exists by that name!")
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

func (c *Client) Access(name string, write bool, a int, data uint64) (uint64, error) {
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

	// write nondummy blocks into stash, record which indexes they're at
	cur_stash := c.stash[name]
	nondummy := find_nondummy(buckets, key)
	path_start := len(cur_stash)
	cur_stash = append(cur_stash, nondummy...)
	path_end := len(cur_stash)
	// fmt.Println("path_start:", path_start)
	// fmt.Println("path_end:", path_end)

	// fmt.Println("nondummy blocks:", nondummy)

	// find index of block we're looking for
	i := slice_find_block(cur_stash, a)

	// modify contents of block in the stash for a write operation
	if write {
		// fmt.Println("if writing, found elem at idx", i)
		new_blk := block_encode(a, data)
		// if element not found, add it as a stash block
		if i == -1 {
			cur_stash = append(cur_stash, new_blk)
		} else {
			cur_stash[i] = new_blk
		}

		ret = data
	} else {
		if i == -1 {
			ret = 0;
		} else {
			_, ret, _ = block_decode(cur_stash[i])
		}
	}

	// fmt.Println("current stash after writing nondummy blocks:", cur_stash)

	// find intersections between old and new path
	old_path, err := s.get_path(x)
	if err != nil {
		return ret, err
	}
	// fmt.Println("old path:", old_path)

	new_path, err := s.get_path(new_leaf)
	if err != nil {
		return ret, err
	}
	// fmt.Println("new path:", new_path)

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
	// fmt.Println("Found intersection:", inters)

	// write back path

	// write back nondummy blocks first
	blk_to_write := make([]Block, 0, len(inters))
	num := cap(blk_to_write) - len(blk_to_write)
	if path_end - path_start < num  {
		num = path_end - path_start
	}

	// fmt.Println("len of blk to write:", len(blk_to_write))
	// fmt.Println("cap of blk to write:", cap(blk_to_write))

	blk_to_write = append(blk_to_write, cur_stash[path_start : path_start + num]...)
	// fmt.Println("Writing back nondummy blocks:", blk_to_write)

	// if there aren't enough blocks to write back, then write stash blocks
	if len(blk_to_write) < cap(blk_to_write) {
		num = cap(blk_to_write) - len(blk_to_write)
		if len(cur_stash) < num {
			num = len(cur_stash)
		}

		blk_to_write = append(blk_to_write, cur_stash[:num]...)
		cur_stash = append(cur_stash[num:])
		// fmt.Printf("Appended %d stash block...\n", num)
	}

	// if out of stash blocks, write dummy blocks
	if len(blk_to_write) < cap(blk_to_write) {
		dummy := make([]Block, len(inters)-len(blk_to_write))
		for i := range dummy {
			dummy[i] = dummy_block()
		}

		blk_to_write = append(blk_to_write, dummy...)
		// fmt.Printf("Appending %d dummy blocks...\n", len(dummy))
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

		// fmt.Printf("Adding %d blocks to stash...\n", len(extra_blks))
	}

	c.stash[name] = cur_stash

	return ret, nil
}
