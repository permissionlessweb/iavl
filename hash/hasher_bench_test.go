package hash

import (
	"crypto/sha256"
	"strconv"
	"testing"

	"github.com/zeebo/blake3"
)

var hashSink []byte

func benchPayload(size int) []byte {
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = byte(i)
	}
	return buf
}

func BenchmarkHashValue(b *testing.B) {
	sizes := []int{32, 64, 256, 512}
	hashers := []struct {
		name   string
		hasher Hasher
	}{
		{"sha256", NewSHA256Hasher()},
		{"blake3", NewBLAKE3Hasher()},
	}
	for _, size := range sizes {
		payload := benchPayload(size)
		for _, h := range hashers {
			h := h
			b.Run(h.name+"/"+strconv.Itoa(size), func(b *testing.B) {
				b.SetBytes(int64(size))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					hashSink = h.hasher.HashValue(payload)
				}
			})
		}
	}
}

func BenchmarkDigestSum256(b *testing.B) {
	sizes := []int{32, 64, 256, 512}
	for _, size := range sizes {
		payload := benchPayload(size)
		b.Run("sha256/"+strconv.Itoa(size), func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sum := sha256.Sum256(payload)
				hashSink = sum[:]
			}
		})
		b.Run("blake3/"+strconv.Itoa(size), func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sum := blake3.Sum256(payload)
				hashSink = sum[:]
			}
		})
	}
}

func BenchmarkIAVLHasherOps(b *testing.B) {
	key := []byte("benchmark-key")
	value := benchPayload(64)
	left := make([]byte, HashSize)
	right := make([]byte, HashSize)

	hashers := []struct {
		name   string
		hasher Hasher
	}{
		{"sha256", NewSHA256Hasher()},
		{"blake3", NewBLAKE3Hasher()},
	}

	for _, h := range hashers {
		h := h
		b.Run(h.name+"/HashLeaf", func(b *testing.B) {
			b.ReportAllocs()
			valueHash := h.hasher.HashValue(value)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hashSink = h.hasher.HashLeaf(0, 1, int64(i), key, valueHash)
			}
		})
		b.Run(h.name+"/HashInner", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				hashSink = h.hasher.HashInner(1, 2, int64(i), left, right)
			}
		})
	}
}
