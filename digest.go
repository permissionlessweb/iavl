package iavl

import (
	"crypto/sha256"
	"hash"

	"github.com/zeebo/blake3"
)

func newDigest(useBlake3 bool) hash.Hash {
	if useBlake3 {
		return blake3.New()
	}
	return sha256.New()
}

func sum256(useBlake3 bool, bz []byte) [32]byte {
	if useBlake3 {
		return blake3.Sum256(bz)
	}
	return sha256.Sum256(bz)
}

func emptyDigest(useBlake3 bool) []byte {
	return newDigest(useBlake3).Sum(nil)
}
