package oram2pc

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"regexp"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type estuple struct {
	entropy        float64
	alphanum_ratio float64
	plain          string
}

// convert a hexstring toa base64 string
func Hex2b64(hexstr string) string {
	bytearray, err := hex.DecodeString(hexstr)
	if err != nil {
		log.Fatal(err)
	}

	b64str := base64.StdEncoding.EncodeToString(bytearray)

	return b64str
}

func XOR_bytes(a []byte, b []byte) []byte {
	if len(a) != len(b) {
		log.Fatal("XOR_bytes: arguments are not the same length!")
	}

	r := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		r[i] = a[i] ^ b[i]
	}

	return r
}

// xor two hexstrings together and return it
func XOR_hex(a string, b string) string {
	a_bytes, err := hex.DecodeString(a)
	if err != nil {
		log.Fatal(err)
	}

	b_bytes, err := hex.DecodeString(b)
	if err != nil {
		log.Fatal(err)
	}

	r_bytes := XOR_bytes(a_bytes, b_bytes)
	return hex.EncodeToString(r_bytes)
}

func Solve_char_xor(a string) {
	a_bytes, err := hex.DecodeString(a)
	if err != nil {
		log.Fatal(err)
	}

	key := make([]byte, len(a_bytes))
	est_map := make(map[byte]estuple)

	for i := 0; i < 256; i++ {
		// build byte array of i's that matches the length
		for j := 0; j < len(key); j++ {
			key[j] = byte(i)
		}

		// decrypt string
		plain_bytes := XOR_bytes(a_bytes, key)
		plain := string(plain_bytes)

		// calculate shannon entropy of the decrypted string
		var entropy float64 = 0.0
		for k := 0; k < 256; k++ {
			k_slice := make([]byte, 1)
			k_slice[0] = byte(k)
			p := bytes.Count(plain_bytes, k_slice)
			p_float := float64(p) / float64(len(plain_bytes))
			if p_float > 0 {
				entropy = entropy - p_float*math.Log2(p_float)
			}
		}
		entropy = entropy / 8.0

		// calculate ratio of alphanumeric chars to other chars
		reg, err := regexp.Compile("[^a-zA-Z0-9]+")
		if err != nil {
			log.Fatal(err)
		}
		processed := reg.ReplaceAllString(plain, "")
		ratio := float64(len(processed)) / float64(len(plain))

		// insert in map
		est := estuple{entropy: entropy, alphanum_ratio: ratio, plain: plain}
		est_map[byte(i)] = est
	}

	// do stats
	var likely_key byte = 0
	most_alphanum := estuple{entropy: 1.0, alphanum_ratio: 0.0, plain: "asdf"}

	for k, v := range est_map {
		if v.alphanum_ratio > most_alphanum.alphanum_ratio {
			most_alphanum = v
			likely_key = k
		}
	}

	fmt.Printf("likely key: %b\n", likely_key)
	fmt.Printf("decrypted to: %s\n", most_alphanum.plain)
	fmt.Printf("with entropy: %f\n", most_alphanum.entropy)
	fmt.Printf("and alphanum ratio: %f\n", most_alphanum.alphanum_ratio)
}

func RandomPerm(size int64) []int64 {
	p := make([]int64, size, size)

	var i int64
	for i = 0; i < size; i++ {
		p[i] = i
	}

	// use fisher-yates shuffling
	for i = 0; i < size - 2; i++ {
		j_big, err := rand.Int(rand.Reader, big.NewInt(size - 1 - i))
		if err != nil {
			log.Println(err)
		}
		j := j_big.Int64()

		p[i], p[j] = p[j], p[i]
	}

	return p
}

// generate a uint64 in [0, max)
func GenUint64(max uint64) uint64 {
	for {
		// get random bytes
		max_f := float64(max - 1)
		num_bits := uint(math.Ceil(math.Log2(max_f)))
		size := uint(math.Ceil(float64(num_bits) / 8.0))
		r := make([]byte, size)
		_, err := rand.Read(r)
		if err != nil {
			log.Println(err)
		}

		// trim bytes to get the right number of bits
		var extra_bits uint
		extra_bits = size * 8 - num_bits
		r[0] = r[0] >> extra_bits

		guess, num := binary.Uvarint(r)
		if num <= 0 {
			continue
		}

		if guess < max {
			return guess
		}
	}
}

func GenAlphanumString(size uint8) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[GenUint64(uint64(len(letters)))]
	}

	return string(b)
}

