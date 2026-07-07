package auth

import (
	"crypto/sha1"
	"fmt"
	"math/big"
)

func minecraftHexDigest(sum []byte) string {
	n := new(big.Int).SetBytes(sum)
	if len(sum) > 0 && sum[0]&0x80 != 0 {
		n.Sub(n, new(big.Int).Lsh(big.NewInt(1), uint(len(sum)*8)))
	}
	return fmt.Sprintf("%x", n)
}

func authDigest(serverID string, sharedSecret, publicKey []byte) string {
	h := sha1.New()
	h.Write([]byte(serverID))
	h.Write(sharedSecret)
	h.Write(publicKey)
	return minecraftHexDigest(h.Sum(nil))
}
