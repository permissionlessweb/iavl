package benchmarks

import (
	"crypto"
	"fmt"
	"hash"
	"testing"

	"github.com/cosmos/iavl"
	iavlhash "github.com/cosmos/iavl/hash"
	"github.com/stretchr/testify/require"

	_ "crypto/sha256"

	_ "golang.org/x/crypto/ripemd160" // nolint: staticcheck // need to test ripemd160
	_ "golang.org/x/crypto/sha3"
)

func BenchmarkHash(b *testing.B) {
	fmt.Printf("%s\n", iavl.GetVersionInfo())

	// Standard hash.Hash implementations
	hashers := []struct {
		name string
		size int
		hash hash.Hash
	}{
		{"ripemd160", 64, crypto.RIPEMD160.New()},
		{"ripemd160", 512, crypto.RIPEMD160.New()},
		{"sha2-256", 64, crypto.SHA256.New()},
		{"sha2-256", 512, crypto.SHA256.New()},
		{"sha3-256", 64, crypto.SHA3_256.New()},
		{"sha3-256", 512, crypto.SHA3_256.New()},
	}

	for _, h := range hashers {
		prefix := fmt.Sprintf("%s-%d", h.name, h.size)
		hasher := h
		b.Run(prefix, func(sub *testing.B) {
			benchHasher(sub, hasher.hash, hasher.size)
		})
	}

	// IAVL hash.Hasher implementations
	iavlHashers := []struct {
		name   string
		size   int
		hasher iavlhash.Hasher
	}{
		{"iavl-sha256", 64, iavlhash.NewSHA256Hasher()},
		{"iavl-sha256", 512, iavlhash.NewSHA256Hasher()},
		{"iavl-poseidon", 64, iavlhash.NewPoseidonHasher()},
		{"iavl-poseidon", 512, iavlhash.NewPoseidonHasher()},
	}

	for _, h := range iavlHashers {
		prefix := fmt.Sprintf("%s-%d", h.name, h.size)
		hasher := h
		b.Run(prefix, func(sub *testing.B) {
			benchIAVLHasher(sub, hasher.hasher, hasher.size)
		})
	}
}

func benchHasher(b *testing.B, hash hash.Hash, size int) {
	// create all random bytes before to avoid timing this
	inputs := randBytes(b.N + size + 1)

	for i := 0; i < b.N; i++ {
		hash.Reset()
		// grab a slice of size bytes from random string
		_, err := hash.Write(inputs[i : i+size])
		require.NoError(b, err)
		hash.Sum(nil)
	}
}

func benchIAVLHasher(b *testing.B, hasher iavlhash.Hasher, size int) {
	// create all random bytes before to avoid timing this
	inputs := randBytes(b.N + size + 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// grab a slice of size bytes from random string
		hasher.HashValue(inputs[i : i+size])
	}
}
