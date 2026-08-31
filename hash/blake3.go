package hash

import (
	"encoding/binary"

	"github.com/zeebo/blake3"
)

// BLAKE3Hasher implements Hasher using BLAKE3-256.
// Node preimages match the original IAVL SHA-256 encoding (varints +
// length-prefixed children) so this is a drop-in hash primitive swap.
type BLAKE3Hasher struct{}

var blake3EmptyHash = func() []byte {
	sum := blake3.Sum256(nil)
	out := make([]byte, HashSize)
	copy(out, sum[:])
	return out
}()

// NewBLAKE3Hasher creates a BLAKE3Hasher.
func NewBLAKE3Hasher() *BLAKE3Hasher {
	return &BLAKE3Hasher{}
}

func blake3Sum(preimage []byte) []byte {
	sum := blake3.Sum256(preimage)
	out := make([]byte, HashSize)
	copy(out, sum[:])
	return out
}

func appendVarint(dst []byte, v int64) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], v)
	return append(dst, buf[:n]...)
}

func appendLenPrefixed(dst, p []byte) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], uint64(len(p)))
	dst = append(dst, buf[:n]...)
	return append(dst, p...)
}

func encodeNodePreimage(dst []byte, height int8, size, version int64, a, b []byte) []byte {
	dst = appendVarint(dst, int64(height))
	dst = appendVarint(dst, size)
	dst = appendVarint(dst, version)
	dst = appendLenPrefixed(dst, a)
	return appendLenPrefixed(dst, b)
}

// HashLeaf computes the BLAKE3 hash of a leaf node.
// Preimage is encoded then hashed with one-shot Sum256, the fastest
// BLAKE3 path for typical IAVL node sizes.
func (h *BLAKE3Hasher) HashLeaf(height int8, size int64, version int64, key []byte, valueHash []byte) []byte {
	preimage := make([]byte, 0, 32+len(key)+len(valueHash))
	preimage = encodeNodePreimage(preimage, height, size, version, key, valueHash)
	return blake3Sum(preimage)
}

// HashInner computes the BLAKE3 hash of an inner node.
func (h *BLAKE3Hasher) HashInner(height int8, size int64, version int64, leftHash []byte, rightHash []byte) []byte {
	var scratch [128]byte
	n := encodeNodePreimage(scratch[:0], height, size, version, leftHash, rightHash)
	return blake3Sum(n)
}

// HashValue computes the BLAKE3 hash of a value.
func (h *BLAKE3Hasher) HashValue(value []byte) []byte {
	return blake3Sum(value)
}

// EmptyHash returns the BLAKE3 hash of an empty input.
func (h *BLAKE3Hasher) EmptyHash() []byte {
	result := make([]byte, len(blake3EmptyHash))
	copy(result, blake3EmptyHash)
	return result
}

// Algorithm returns BLAKE3.
func (h *BLAKE3Hasher) Algorithm() HashAlgorithm {
	return BLAKE3
}
