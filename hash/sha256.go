package hash

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
)

// SHA256Hasher implements Hasher with SHA-256 and the original IAVL
// writeHashBytes encoding (varints + length-prefixed children).
type SHA256Hasher struct{}

var (
	sha256HasherPool = &sync.Pool{
		New: func() any {
			return sha256.New()
		},
	}
	sha256EmptyHash = sha256.New().Sum(nil)
)

// NewSHA256Hasher creates a SHA256Hasher.
func NewSHA256Hasher() *SHA256Hasher {
	return &SHA256Hasher{}
}

type sha256Digest interface {
	Write([]byte) (int, error)
	Sum([]byte) []byte
	Reset()
}

func getSHA256() sha256Digest {
	h := sha256HasherPool.Get().(sha256Digest)
	h.Reset()
	return h
}

func putSHA256(h sha256Digest) {
	h.Reset()
	sha256HasherPool.Put(h)
}

// HashLeaf computes the SHA-256 hash of a leaf node.
func (h *SHA256Hasher) HashLeaf(height int8, size int64, version int64, key []byte, valueHash []byte) []byte {
	hasher := getSHA256()
	defer putSHA256(hasher)

	var buf [binary.MaxVarintLen64]byte

	n := binary.PutVarint(buf[:], int64(height))
	hasher.Write(buf[:n])

	n = binary.PutVarint(buf[:], size)
	hasher.Write(buf[:n])

	n = binary.PutVarint(buf[:], version)
	hasher.Write(buf[:n])

	n = binary.PutUvarint(buf[:], uint64(len(key)))
	hasher.Write(buf[:n])
	hasher.Write(key)

	n = binary.PutUvarint(buf[:], uint64(len(valueHash)))
	hasher.Write(buf[:n])
	hasher.Write(valueHash)

	return hasher.Sum(nil)
}

// HashInner computes the SHA-256 hash of an inner node.
func (h *SHA256Hasher) HashInner(height int8, size int64, version int64, leftHash []byte, rightHash []byte) []byte {
	hasher := getSHA256()
	defer putSHA256(hasher)

	var buf [binary.MaxVarintLen64]byte

	n := binary.PutVarint(buf[:], int64(height))
	hasher.Write(buf[:n])

	n = binary.PutVarint(buf[:], size)
	hasher.Write(buf[:n])

	n = binary.PutVarint(buf[:], version)
	hasher.Write(buf[:n])

	n = binary.PutUvarint(buf[:], uint64(len(leftHash)))
	hasher.Write(buf[:n])
	hasher.Write(leftHash)

	n = binary.PutUvarint(buf[:], uint64(len(rightHash)))
	hasher.Write(buf[:n])
	hasher.Write(rightHash)

	return hasher.Sum(nil)
}

// HashValue computes the SHA-256 hash of a value.
func (h *SHA256Hasher) HashValue(value []byte) []byte {
	sum := sha256.Sum256(value)
	out := make([]byte, HashSize)
	copy(out, sum[:])
	return out
}

// EmptyHash returns the SHA-256 hash of an empty input.
func (h *SHA256Hasher) EmptyHash() []byte {
	result := make([]byte, len(sha256EmptyHash))
	copy(result, sha256EmptyHash)
	return result
}

// Algorithm returns SHA256.
func (h *SHA256Hasher) Algorithm() HashAlgorithm {
	return SHA256
}
