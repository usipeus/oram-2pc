/*
 * Implements various functions related to blocks
 */

package oram2pc

import (
	"bytes"
	"fmt"
	"encoding/binary"
)

type Block []byte

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
