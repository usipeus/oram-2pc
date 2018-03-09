/*
 * Test functions for the oram2pc library
 */
package oram2pc

import (
	"fmt"
	"strconv"
	"testing"
)

func Test_blocks(t *testing.T) {
	key := []byte("key lul")

	aoeu := enc_dummy_block(key)
	aoeu2 := enc_dummy_block(key)
	aoeu3 := enc_dummy_block(key)
	aoeu4 := enc_dummy_block(key)
	b := enc_block(block_encode(0x1234, 0x1122334455667788), key)
	fmt.Println(block_encode(0x1234, 0x1122334455667788))
	fmt.Println(b)
	id, d, dummy := block_decode(dec_block(b, key))
	if dummy == false {
		fmt.Printf("%x: %x\n", id, d)
	}

	_, e, dummy := block_decode(dec_block(aoeu, key))
	if dummy == true {
		fmt.Println("found dummy block!", e)
	}

	bucket := Bucket{b, aoeu2, aoeu3, aoeu4}
	dummy_bucket := Bucket{aoeu, aoeu2, aoeu3, aoeu4}
	joined := bucket_join(bucket, nil)
	fmt.Println(aoeu)
	fmt.Println(joined[:32])
	fmt.Println(aoeu2)
	fmt.Println(joined[32:64])
	fmt.Println(aoeu3)
	fmt.Println(joined[64:96])
	fmt.Println(aoeu4)
	fmt.Println(joined[96:])
	if len(joined) != len(bucket[0])*4 {
		panic("Check bucket_join!!!")
	}

	bux := []Bucket{dummy_bucket, dummy_bucket, dummy_bucket, bucket}
	fmt.Println(find_nondummy(bux, key))

	idx, val := find_block(bux, 0x1234, key)
	fmt.Println("Finding nondummy in buckets: index", idx, "val", val)

	bucket2 := make_bucket([]Block{aoeu}, 4, key)
	fmt.Println(bucket2)
}

func Test_client(t *testing.T) {
	// currently use 256-bit blocks to store encrypted 64-bit values
	c := InitClient(4, 4)

	c.AddServer("test", c.N, c.Z, 4096)
	fmt.Println(c.ServerInfo("test"))

	s, _ := c.servers["test"]
	buckets, err := s.get_path_buckets(2)
	if err != nil {
		panic(err)
	}

	for i := range buckets {
		fmt.Println("Reading bucket", strconv.Itoa(i))
		for j := range buckets[i] {
			fmt.Println("Reading block", strconv.Itoa(j))
			fmt.Println(buckets[i][j])
		}
	}

	fmt.Println("Trying to write to a = 0")
	_, err = c.Access("test", true, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println("Trying to read from a = 0")
	val, err := c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
	fmt.Println("Trying to read from a = 0")
	val, err = c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
	fmt.Println("Trying to read from a = 0")
	val, err = c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
	fmt.Println("Trying to read from a = 0")
	val, err = c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
	fmt.Println("Trying to write to a = 1")
	_, err = c.Access("test", true, 1, uint64(0x10))
	if err != nil {
		panic(err)
	}
	fmt.Println("Trying to read from a = 0")
	val, err = c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)

	c.RemoveServer("test")
}
