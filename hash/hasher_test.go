package hash

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

const (
	sha256EmptyHex = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	// Official BLAKE3 test vector for the empty input (32-byte output).
	// https://github.com/BLAKE3-team/BLAKE3/blob/master/test_vectors/test_vectors.json
	blake3EmptyHex = "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"
)

func TestSHA256EmptyHash(t *testing.T) {
	h := NewSHA256Hasher()
	if h.Algorithm() != SHA256 {
		t.Errorf("algorithm: got %s want %s", h.Algorithm(), SHA256)
	}
	want, err := hex.DecodeString(sha256EmptyHex)
	if err != nil {
		t.Fatal(err)
	}
	got := h.EmptyHash()
	if !bytes.Equal(got, want) {
		t.Errorf("empty hash:\n got %x\nwant %x", got, want)
	}
	if !bytes.Equal(h.HashValue(nil), want) {
		t.Error("HashValue(nil) should equal EmptyHash")
	}
	got[0] ^= 0xff
	if bytes.Equal(h.EmptyHash(), got) {
		t.Error("EmptyHash returned a shared mutable buffer")
	}
}

func TestBLAKE3EmptyHash(t *testing.T) {
	h := NewBLAKE3Hasher()
	if h.Algorithm() != BLAKE3 {
		t.Errorf("algorithm: got %s want %s", h.Algorithm(), BLAKE3)
	}
	want, err := hex.DecodeString(blake3EmptyHex)
	if err != nil {
		t.Fatal(err)
	}
	got := h.EmptyHash()
	if !bytes.Equal(got, want) {
		t.Errorf("empty hash:\n got %x\nwant %x", got, want)
	}
	if !bytes.Equal(h.HashValue(nil), want) {
		t.Error("HashValue(nil) should equal EmptyHash")
	}
	got[0] ^= 0xff
	if bytes.Equal(h.EmptyHash(), got) {
		t.Error("EmptyHash returned a shared mutable buffer")
	}
}

func TestHashValueDeterminism(t *testing.T) {
	value := []byte("hello world")
	for _, h := range []Hasher{NewSHA256Hasher(), NewBLAKE3Hasher()} {
		a := h.HashValue(value)
		b := h.HashValue(value)
		if !bytes.Equal(a, b) {
			t.Errorf("%s HashValue is not deterministic", h.Algorithm())
		}
		if len(a) != HashSize {
			t.Errorf("%s HashValue length %d", h.Algorithm(), len(a))
		}
		if bytes.Equal(a, h.HashValue([]byte("different"))) {
			t.Errorf("%s different values produced same hash", h.Algorithm())
		}
	}

	sum := sha256.Sum256(value)
	if !bytes.Equal(NewSHA256Hasher().HashValue(value), sum[:]) {
		t.Error("SHA256Hasher.HashValue must match crypto/sha256.Sum256")
	}
}

func TestHashLeafDeterminism(t *testing.T) {
	key := []byte("test-key")
	for _, h := range []Hasher{NewSHA256Hasher(), NewBLAKE3Hasher()} {
		valueHash := h.HashValue([]byte("test-value"))
		a := h.HashLeaf(0, 1, 1, key, valueHash)
		b := h.HashLeaf(0, 1, 1, key, valueHash)
		if !bytes.Equal(a, b) {
			t.Errorf("%s HashLeaf is not deterministic", h.Algorithm())
		}
		if len(a) != HashSize {
			t.Errorf("%s HashLeaf length %d", h.Algorithm(), len(a))
		}
		if bytes.Equal(a, h.HashLeaf(0, 1, 1, []byte("other"), valueHash)) {
			t.Errorf("%s different keys produced same leaf hash", h.Algorithm())
		}
		if bytes.Equal(a, h.HashLeaf(0, 1, 2, key, valueHash)) {
			t.Errorf("%s different versions produced same leaf hash", h.Algorithm())
		}
	}
}

func TestHashInnerDeterminism(t *testing.T) {
	for _, h := range []Hasher{NewSHA256Hasher(), NewBLAKE3Hasher()} {
		left := h.HashValue([]byte("left"))
		right := h.HashValue([]byte("right"))
		a := h.HashInner(1, 2, 1, left, right)
		b := h.HashInner(1, 2, 1, left, right)
		if !bytes.Equal(a, b) {
			t.Errorf("%s HashInner is not deterministic", h.Algorithm())
		}
		if bytes.Equal(a, h.HashInner(1, 2, 1, right, left)) {
			t.Errorf("%s swapped children produced same inner hash", h.Algorithm())
		}
		if bytes.Equal(a, h.HashInner(2, 2, 1, left, right)) {
			t.Errorf("%s different heights produced same inner hash", h.Algorithm())
		}
	}
}

func TestBLAKE3DifferentFromSHA256(t *testing.T) {
	blake3Hasher := NewBLAKE3Hasher()
	sha256Hasher := NewSHA256Hasher()
	value := []byte("test value")
	if bytes.Equal(blake3Hasher.HashValue(value), sha256Hasher.HashValue(value)) {
		t.Error("BLAKE3 and SHA256 should produce different hashes")
	}
	if bytes.Equal(blake3Hasher.EmptyHash(), sha256Hasher.EmptyHash()) {
		t.Error("empty hashes should differ")
	}

	key := []byte("k")
	vhB := blake3Hasher.HashValue(value)
	vhS := sha256Hasher.HashValue(value)
	if bytes.Equal(blake3Hasher.HashLeaf(0, 1, 1, key, vhB), sha256Hasher.HashLeaf(0, 1, 1, key, vhS)) {
		t.Error("leaf hashes should differ")
	}
}
