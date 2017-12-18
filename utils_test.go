package oram2pc

import (
	"fmt"
	"crypto/rand"
	"testing"
)

func test_cryptopals(t *testing.T) {
	a := "49276d206b696c6c696e6720796f757220627261696e206c696b65206120706f69736f6e6f7573206d757368726f6f6d"
	b := Hex2b64(a)
	fmt.Printf("a: %s\n", a)
	fmt.Printf("b: %s\n", b)

	c := "1c0111001f010100061a024b53535009181c"
	d := "686974207468652062756c6c277320657965"
	fmt.Printf("%s\nXOR\n%s is:\n%s\n", c, d, XOR_hex(c, d))

	e := "1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736"
	Solve_char_xor(e)
}

func Test_utils(t *testing.T) {
	m := make([]byte, 4)
	rand.Read(m)
	fmt.Println("Random test subject:", m)

	fmt.Println("Testing pad functions:")
	fmt.Println(pad(m, 128/8, 0x24))
	fmt.Println(unpad(pad(m, 128/8, 0x24), 0x24))

	fmt.Println("Testing plaintext encoding and decoding functions:")
	fmt.Println(pt_encode(m, 128/8))
	fmt.Println(pt_decode(pt_encode(m, 128/8)))

	k := make([]byte, 128/8)
	rand.Read(k)
	cip := Encrypt(m, k)
	fmt.Println("Testing encrypting and decrypting...")
	fmt.Println(cip)
	fmt.Println(Decrypt(cip, k))
}

