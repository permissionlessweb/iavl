package iavl

import (
	"crypto/sha256"
	"hash"
	"sync"

	"github.com/zeebo/blake3"
	"golang.org/x/crypto/blake2b"
)

// HashAlgo selects the 32-byte node digest. Default is SHA-256 (ICS23 IavlSpec).
type HashAlgo uint8

const (
	HashSHA256 HashAlgo = iota
	HashBLAKE3
	HashBLAKE2b256
)

func (a HashAlgo) String() string {
	switch a {
	case HashBLAKE3:
		return "blake3"
	case HashBLAKE2b256:
		return "blake2b256"
	default:
		return "sha256"
	}
}

var (
	sha256Pool = sync.Pool{New: func() any { return sha256.New() }}
	blake3Pool = sync.Pool{New: func() any { return blake3.New() }}
	blake2b256Pool = sync.Pool{New: func() any {
		h, err := blake2b.New256(nil)
		if err != nil {
			panic(err)
		}
		return h
	}}
)

func digestPool(algo HashAlgo) *sync.Pool {
	switch algo {
	case HashBLAKE3:
		return &blake3Pool
	case HashBLAKE2b256:
		return &blake2b256Pool
	default:
		return &sha256Pool
	}
}

func newDigest(algo HashAlgo) hash.Hash {
	switch algo {
	case HashBLAKE3:
		return blake3.New()
	case HashBLAKE2b256:
		h, err := blake2b.New256(nil)
		if err != nil {
			panic(err)
		}
		return h
	default:
		return sha256.New()
	}
}

func getDigest(algo HashAlgo) hash.Hash {
	return digestPool(algo).Get().(hash.Hash)
}

func putDigest(algo HashAlgo, h hash.Hash) {
	h.Reset()
	digestPool(algo).Put(h)
}

func sum256(algo HashAlgo, bz []byte) [32]byte {
	switch algo {
	case HashBLAKE3:
		return blake3.Sum256(bz)
	case HashBLAKE2b256:
		h, err := blake2b.New256(nil)
		if err != nil {
			panic(err)
		}
		_, _ = h.Write(bz)
		var out [32]byte
		copy(out[:], h.Sum(nil))
		return out
	default:
		return sha256.Sum256(bz)
	}
}

func emptyDigest(algo HashAlgo) []byte {
	return newDigest(algo).Sum(nil)
}
