package protocol

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"testing"
)

func TestCFB8_NISTVector(t *testing.T) {
	key, _ := hex.DecodeString("2b7e151628aed2a6abf7158809cf4f3c")
	iv, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	plain, _ := hex.DecodeString("6bc1bee22e409f96e93d7e117393172a")
	want, _ := hex.DecodeString("3b79424c9c0dd436bace9e0ed4586a4f")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	got := make([]byte, len(plain))
	NewCFB8Encrypter(block, iv).XORKeyStream(got, plain)
	if !bytes.Equal(got, want) {
		t.Errorf("encrypt = %x, want %x", got, want)
	}

	back := make([]byte, len(want))
	NewCFB8Decrypter(block, iv).XORKeyStream(back, got)
	if !bytes.Equal(back, plain) {
		t.Errorf("decrypt = %x, want %x", back, plain)
	}
}
