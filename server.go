/*
 * Implementation of the server in Path ORAM by Stefanov et. al.
 */
package oram2pc

import (
	"errors"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

/*
 * Each bucket contains Z blocks; it suffices that Z is small e.g. 4
 */
type Bucket []Block

/*
 * The server
 */
type Server struct {
	N     int    // total number of blocks outsourced
	L     int    // height of binary tree
	B     int    // block size in bytes, currently fixed at 32
	Z     int    // capacity of each bucket in blocks
	dir   string // directory that holds the tree, stored as files
	fsize int    // filesize of each file that represents a level
}

/*
 * Initialize a Server
 *
 * Returns a new Server with params:
 *   N: Number of blocks outsourced to the server
 *   B: Capacity of each block in bytes (fixed to 32 bytes)
 *   Z: Capacity of each bucket in blocks
 */
func init_server(N int, Z int, fsize int) *Server {
	s := &Server{N: N, B: 32, Z: Z}
	s.dir = filepath.Join(os.TempDir(), gen_alphanum_string(10))
	// height of tree: log2(N)
	s.L = int(math.Ceil(math.Log2(float64(N))))

	s.fsize = fsize

	return s
}

// get full path to the nth bucket on the lth level
func (s *Server) get_fp(l int, n int) string {
	fname := filepath.Join(s.dir, strconv.Itoa(l)+"."+strconv.Itoa(n))
	return fname
}

func (s *Server) create_tree() {
	// create directory
	os.Mkdir(s.dir, 0755)

	// write each file with all zeroes
	buf := make([]byte, s.fsize)
	for i := 0; i <= s.L; i++ {
		// for each level of the tree, create at least 1 file
		lvl_bytes := (1 << uint(i)) * s.Z * s.B

		for j := 0; j <= (lvl_bytes / s.fsize); j++ {
			fp := s.get_fp(i, j)

			err := ioutil.WriteFile(fp, buf, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (s *Server) write_node(b Bucket, l int, n int) {
	max_n := (1 << uint(l))
	if n < 0 || n >= max_n {
		return
	}

	// get raw bytes of bucket
	bucket_bytes := bucket_join(b, nil)

	fp, off := s.foffset(l, n)
	f, err := os.OpenFile(fp, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	_, err = f.WriteAt(bucket_bytes, int64(off))
	if err != nil {
		panic(err)
	}

	err = f.Sync()
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}
}

func (s *Server) read_node(l int, n int) (Bucket, error) {
	// get file and offset into that file
	fp, offset := s.foffset(l, n)

	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}

	// in bytes
	bucket_size := s.B * s.Z

	// read all bytes at once
	buf := make([]byte, bucket_size)
	m, err := f.ReadAt(buf, int64(offset))
	if m < len(buf) || err != nil {
		return nil, err
	}

	// organize bytes into buckets
	bucket := make(Bucket, s.Z)
	for i := range bucket {
		left := i * s.B
		right := (i + 1) * s.B
		bucket[i] = buf[left:right]
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

// returns the path to block n
func (s *Server) get_path(n int) ([]int, error) {
	if n < 0 || n >= s.N {
		// block not found
		return nil, errors.New("Block number out of range")
	}

	// for each level of the tree, get which index the bucket is
	path := make([]int, s.L+1)

	cur_n := n
	for i := s.L; i >= 0; i-- {
		path[i] = cur_n
		cur_n /= 2
	}

	return path, nil
}

// returns the file and an offset into that file for a bucket at a given level
func (s *Server) foffset(l int, n int) (string, int) {
	// bounds check
	if n < 0 || n >= (1<<uint(l)) {
		return "", 0
	}

	total_bytes := s.B * s.Z * n
	off := total_bytes % s.fsize
	fp := s.get_fp(l, total_bytes/s.fsize)

	return fp, off
}

// access the files stored on disk to retrieve buckets of a path
func (s *Server) get_path_buckets(n int) ([]Bucket, error) {
	path, err := s.get_path(n)
	if err != nil {
		return nil, err
	}

	bux := make([]Bucket, s.L+1)
	for i := range bux {
		bucket, err := s.read_node(i, path[i])
		if err != nil {
			return nil, err
		}

		bux[i] = bucket
	}

	return bux, nil
}
