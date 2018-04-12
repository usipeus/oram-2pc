/*
 * Implements various functions related to blocks
 */

package oram2pc

import (
	"bytes"
	"encoding/binary"
)

type Block []byte

// makes a bucket consisting of blks and padded with dummy blocks until
// the bucket reaches a size of max
func make_bucket(blks []Block, max int, key []byte) Bucket {

	bucket := make(Bucket, max)

	end := len(blks)
	if end > max {
		end = max;
	}

	for i := 0; i < end; i++ {
		bucket[i] = enc_block(blks[i], key)
	}

	// pad with encrypted dummy blocks
	for i := end; i < max; i++ {
		bucket[i] = enc_dummy_block(key)
	}

	return bucket
}

// splits a bucket into its plaintext blocks (opposite of make_bucket)
func split_bucket(bucket Bucket, key []byte) []Block {
	blocks := make([]Block, len(bucket))
	for i := range bucket {
		blocks[i] = dec_block(bucket[i], key)
	}

	return blocks
}

// finds all non-dummy blocks in some buckets
func find_nondummy(bux []Bucket, key []byte) []Block {
	nondummy := make([]Block, len(bux)*len(bux[0]))
	num_nd := 0
	for i := range bux {
		for j := range bux[i] {
			cur_blk := dec_block(bux[i][j], key)
			_, _, dummy := block_decode(cur_blk)
			if !dummy {
				nondummy[num_nd] = cur_blk
				num_nd += 1
				continue
			}
		}
	}

	return nondummy[:num_nd]
}

// find which bucket the block with id "id" is found
func find_block(bux []Bucket, id int, key []byte) (int, uint64) {
	for i := range bux {
		for j := range bux[i] {
			cur_blk := bux[i][j]
			cur_id, val, dummy := block_decode(dec_block(cur_blk, key))

			if dummy == false && cur_id == id {
				return i, val
			}
		}
	}

	return -1, 0
}

// Join concatenates the elements of s to create a new byte slice. The separator
// sep is placed between elements in the resulting slice.
func bucket_join(s Bucket, sep []byte) []byte {

	if len(s) == 0 {
		return []byte{}
	}

	if len(s) == 1 {
		// Just return a copy.
		return append([]byte(nil), s[0]...)
	}

	n := len(sep) * (len(s) - 1)
	for _, v := range s {
		n += len(v)
	}

	b := make([]byte, n)
	bp := copy(b, s[0])
	for _, v := range s[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], v)
	}

	return b
}

/*
 * Returns an unencrypted dummy block
 */
func dummy_block() Block {
	dummy := make([]byte, 16)
	for i := range dummy {
		dummy[i] = 0xff
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
 * | 0xFFFF... |
 * <- 128 bits ->
 */
func enc_dummy_block(k []byte) Block {
	dummy_plain := dummy_block()
	dummy_cip := encrypt(dummy_plain, k)

	return dummy_cip
}

// parse an id and value to create a plaintext block ready for encryption
func block_encode(id int, val uint64) []byte {
	// left pad with the id, extended to 8 bytes so we know it's not a dummy blk
	id_bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(id_bytes, uint64(id))
	val_bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(val_bytes, val)

	blk_plain := append(id_bytes, val_bytes...)

	return blk_plain
}

// do the opposite of block_encode
func block_decode(blk []byte) (int, uint64, bool) {
	if is_dummy(blk) {
		return 0, uint64(0), true
	}

	id := binary.LittleEndian.Uint64(blk[:8])
	val := binary.LittleEndian.Uint64(blk[8:])
	return int(id), val, false
}

/*
 * Returns an encrypted version of the encoded block
 */
func enc_block(blk []byte, k []byte) []byte {
	return encrypt(blk, k)
}

/*
 * Returns the plaintext (a uint64) by decrypting an encrypted block if
 * the bool == false, else the block is a dummy block
 */
func dec_block(blk []byte, k []byte) []byte {
	return decrypt(blk, k)
}
