package iavl

import (
	"crypto/sha256"
	"hash"
	"sync"

	"github.com/zeebo/blake3"
)

var (
	sha256Pool = sync.Pool{New: func() any { return sha256.New() }}
	blake3Pool = sync.Pool{New: func() any { return blake3.New() }}
)

func digestPool(useBlake3 bool) *sync.Pool {
	if useBlake3 {
		return &blake3Pool
	}
	return &sha256Pool
}

func newDigest(useBlake3 bool) hash.Hash {
	if useBlake3 {
		return blake3.New()
	}
	return sha256.New()
}

func getDigest(useBlake3 bool) hash.Hash {
	return digestPool(useBlake3).Get().(hash.Hash)
}

func putDigest(useBlake3 bool, h hash.Hash) {
	h.Reset()
	digestPool(useBlake3).Put(h)
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
