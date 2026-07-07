package protocol

import (
	"bytes"
	"crypto/cipher"
)

type cfb8 struct {
	block   cipher.Block
	iv      []byte
	tmp     []byte
	decrypt bool
}

func newCFB8(block cipher.Block, iv []byte, decrypt bool) *cfb8 {
	if len(iv) != block.BlockSize() {
		panic("cfb8: length IV must match the block size")
	}
	return &cfb8{
		block:   block,
		iv:      bytes.Clone(iv),
		tmp:     make([]byte, block.BlockSize()),
		decrypt: decrypt,
	}
}

func NewCFB8Encrypter(block cipher.Block, iv []byte) cipher.Stream {
	return newCFB8(block, iv, false)
}
func NewCFB8Decrypter(block cipher.Block, iv []byte) cipher.Stream {
	return newCFB8(block, iv, true)
}

func (c *cfb8) XORKeyStream(dst, src []byte) {
	for i := range src {
		c.block.Encrypt(c.tmp, c.iv)
		k := c.tmp[0]

		in := src[i]
		out := in ^ k
		dst[i] = out

		fed := out
		if c.decrypt {
			fed = in
		}

		copy(c.iv, c.iv[1:])
		c.iv[len(c.iv)-1] = fed
	}
}
