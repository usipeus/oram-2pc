/*
 * Implementation of the server in Path ORAM by Stefanov et. al.
 */
package oram2pc

import (
	"bytes"
	"fmt"
	"encoding/binary"
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
 * Returns an unencrypted dummy block
 */
func dummy_block() Block {
	dummy := make([]byte, 16)
	for i := 0; i < 8; i++ {
		dummy[i] = 0xff
		dummy[8 + i] = 0x00
	}

	return dummy
}

/*
 * Detects whether the byte slice is the unencrypted dummy block
 */
func is_dummy(blk []byte) bool {
	result := bytes.Compare(blk, dummy_block())
	return result == 0
}

/*
 * Returns an encrypted version of the dummy block which has the format:
 * | 0xFFFF... | | 0x0000... |
 * <- 64 bits -> <- 64 bits ->
 */
func enc_dummy_block(k []byte) Block {
	dummy_plain := dummy_block()

	fmt.Println(dummy_plain, len(dummy_plain))
	dummy_cip := Encrypt(dummy_plain, k)
	fmt.Println(dummy_cip, len(dummy_cip))

	return dummy_cip
}

/*
 * Returns an encrypted version of the uint64 value
 */
func enc_block(val uint64, k []byte) []byte {
	// left pad with 8 bytes of 0s so we know it's not a dummy blk
	val_bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(val_bytes, val)

	blk_plain := append(make([]byte, 8), val_bytes...)
	return Encrypt(blk_plain, k)
}

/*
 * Returns the plaintext (an int64) by decrypting an encrypted block, if
 * the bool  = 0, else the block is a dummy block
 */
func dec_block(blk []byte, k []byte) (uint64, bool) {
	blk_plain := Decrypt(blk, k)
	if is_dummy(blk_plain) {
		return uint64(0), true
	}

	val := binary.LittleEndian.Uint64(blk_plain[8:])
	return val, false
}

/*
 * Initialize a Server
 *
 * Returns a new Server with params:
 *   N: Number of blocks outsourced to the server
 *   B: Capacity of each block in bytes (fixed to 32 bytes)
 *   Z: Capacity of each bucket in blocks
 */
func Init_server(N int, Z int, fsize int) *Server {
	s := &Server{N: N, B: 32, Z: Z}
	s.dir = filepath.Join(os.TempDir(), GenAlphanumString(10))
	// height of tree: log2(N)
	s.L = int(math.Ceil(math.Log2(float64(N))))

	s.fsize = fsize

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

