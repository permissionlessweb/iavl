package iavl

import (
	"strconv"
	"testing"
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
	sizes := []int{32, 64, 75, 256}
	for _, size := range sizes {
		payload := benchPayload(size)
		for _, algo := range benchAlgos {
			n := &Node{algo: algo}
			b.Run(algo.String()+"/"+strconv.Itoa(size), func(b *testing.B) {
				b.SetBytes(int64(size))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					sum := n.sum256(payload)
					hashSink = sum[:]
				}
			})
		}
	}
}

func benchV2Leaf(algo HashAlgo) *Node {
	return &Node{
		key:           []byte("benchmark-key-25-bytes!!"),
		value:         benchPayload(64),
		subtreeHeight: 0,
		size:          1,
		nodeKey:       NewNodeKey(1, 1),
		algo:          algo,
	}
}

func benchV2Inner(algo HashAlgo) *Node {
	left := benchV2Leaf(algo)
	left._hash()
	right := benchV2Leaf(algo)
	right.key = []byte("benchmark-key-right!!!!!")
	right.nodeKey = NewNodeKey(1, 2)
	right.hash = nil
	right._hash()
	return &Node{
		key:           right.key,
		subtreeHeight: 1,
		size:          2,
		nodeKey:       NewNodeKey(1, 3),
		leftNode:      left,
		rightNode:     right,
		algo:          algo,
	}
}

func BenchmarkNode_hash(b *testing.B) {
	for _, algo := range benchAlgos {
		b.Run(algo.String()+"/leaf", func(b *testing.B) {
			node := benchV2Leaf(algo)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.hash = nil
				hashSink = node._hash()
			}
		})
		b.Run(algo.String()+"/inner", func(b *testing.B) {
			node := benchV2Inner(algo)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.hash = nil
				hashSink = node._hash()
			}
		})
	}
}
