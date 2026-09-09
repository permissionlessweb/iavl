package iavl

import (
	"strconv"
	"testing"

	dbm "github.com/cosmos/iavl/db"
)

var hashSink []byte

func benchPayload(size int) []byte {
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = byte(i)
	}
	return buf
}

var benchAlgos = []HashAlgo{HashSHA256, HashBLAKE3, HashBLAKE2b256}

func BenchmarkDigestSum256(b *testing.B) {
	// 32 = child hash; 64 = one BLAKE3 block; 75 ≈ inner node preimage (2×64 B blocks).
	sizes := []int{32, 64, 75, 256}
	for _, size := range sizes {
		payload := benchPayload(size)
		for _, algo := range benchAlgos {
			b.Run(algo.String()+"/"+strconv.Itoa(size), func(b *testing.B) {
				b.SetBytes(int64(size))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					sum := sum256(algo, payload)
					hashSink = sum[:]
				}
			})
		}
	}
}

func benchLeaf(algo HashAlgo) *Node {
	n := NewNode([]byte("benchmark-key-25-bytes!!"), benchPayload(64))
	n.algo = algo
	n.nodeKey = &NodeKey{version: 1, nonce: 1}
	return n
}

func benchInner(algo HashAlgo) *Node {
	left := benchLeaf(algo)
	left.hash = left._hash(1)
	right := benchLeaf(algo)
	right.key = []byte("benchmark-key-right!!!!!")
	right.hash = nil
	right.hash = right._hash(1)
	return &Node{
		key:           right.key,
		subtreeHeight: 1,
		size:          2,
		nodeKey:       &NodeKey{version: 1, nonce: 2},
		leftNode:      left,
		rightNode:     right,
		algo:          algo,
	}
}

func BenchmarkNode_hash(b *testing.B) {
	for _, algo := range benchAlgos {
		b.Run(algo.String()+"/leaf", func(b *testing.B) {
			node := benchLeaf(algo)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.hash = nil
				hashSink = node._hash(1)
			}
		})
		b.Run(algo.String()+"/inner", func(b *testing.B) {
			node := benchInner(algo)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.hash = nil
				hashSink = node._hash(1)
			}
		})
	}
}

func BenchmarkMutableTree_SetWorkingHash(b *testing.B) {
	for _, algo := range benchAlgos {
		opts := []Option{}
		switch algo {
		case HashBLAKE3:
			opts = append(opts, Blake3Option())
		case HashBLAKE2b256:
			opts = append(opts, Blake2b256Option())
		}
		b.Run(algo.String(), func(b *testing.B) {
			tree := NewMutableTree(dbm.NewMemDB(), 0, true, NewNopLogger(), opts...)
			key := []byte("k")
			val := benchPayload(32)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				val[0] = byte(i)
				_, err := tree.Set(key, val)
				if err != nil {
					b.Fatal(err)
				}
				hashSink = tree.WorkingHash()
			}
		})
	}
}
