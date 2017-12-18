/*
 * Implementation of the server in Path ORAM by Stefanov et. al.
 */
package oram2pc

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

type Block []byte

/*
 * Each bucket contains Z blocks; it suffices that Z is small e.g. 4
 */
type Bucket []Block

/*
 * The server
 */
type Server struct {
	N int // total number of blocks outsourced
	L int // height of binary tree
	B int // block size in bytes
	Z int // capacity of each bucket in blocks
	dir string // directory that holds the tree, stored as files
	fsize int // filesize of each file that represents a level
}

/*
 * Initialize a Server
 *
 * Returns a new Server with params:
 *   N: Number of blocks outsourced to the server
 *   B: Capacity of each block in bytes
 *   Z: Capacity of each bucket in blocks
 */
func Init_server(N int, B int, Z int, fsize int) *Server {
	s := &Server{N: N, B: B, Z: Z}
	s.dir = filepath.Join(os.TempDir(), GenAlphanumString(10))
	// height of tree: log2(N)
	s.L = int(math.Ceil(math.Log2(float64(N))))

	s.fsize = fsize

	// initialize files
	for i := 0; i <= s.L; i++ {
	}

	return s
}

// returns the path to block n
func (s *Server) get_path(n int) ([]int, error) {
	if n < 0 || n >= s.N {
		// block not found
		return nil, errors.New("Block number out of range")
	}

	// for each level of the tree, get which index the bucket is
	path := make([]int, s.L + 1)

	cur_n := n
	for i := s.L; i >= 0; i-- {
		path[i] = cur_n
		cur_n /= 2
	}

	return path, nil
}

// returns the file and an offset into that file for a bucket at a given level
func (s *Server) foffset(n int, l int) (string, int) {
	// bounds check
	if n < 0 || n > (1 << uint(l)) {
		return "", 0
	}

	byte_offset := (s.Z * n) % s.fsize
	fname := strconv.Itoa(l) + "." + strconv.Itoa(s.Z * n / s.fsize)

	return filepath.Join(s.dir, fname), byte_offset
}

// access the files stored on disk to retrieve buckets of a path
func (s *Server) get_buckets(n int) ([]Bucket, error) {
	path, err := s.get_path(n)
	if err != nil {
		return nil, err
	}

	bux := make([]Bucket, s.L)
	for i := range bux {
		// get file and offset into that file
		fp, offset := s.foffset(i, path[i])

		f, err := os.Open(fp)
		if err != nil {
			return nil, err
		}

		// read all bytes at once
		buf := make([]byte, s.B * s.Z)
		n, err := f.ReadAt(buf, int64(offset))
		if n < len(buf) || err != nil {
			return nil, err
		}

		// organize bytes into buckets
		bux[i] = make(Bucket, s.Z)
		for j := 0; j < s.Z; j++ {
			bux[i][j] = buf[j * s.Z : (j + 1) * s.Z]
		}
	}

	return bux, nil
}

