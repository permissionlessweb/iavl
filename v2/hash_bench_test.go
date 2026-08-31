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

func BenchmarkDigestSum256(b *testing.B) {
	sizes := []int{32, 64, 75, 256}
	for _, size := range sizes {
		payload := benchPayload(size)
		for _, blake := range []bool{false, true} {
			name := "sha256/" + strconv.Itoa(size)
			if blake {
				name = "blake3/" + strconv.Itoa(size)
			}
			n := &Node{useBlake3: blake}
			b.Run(name, func(b *testing.B) {
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

func benchV2Leaf(useBlake3 bool) *Node {
	return &Node{
		key:           []byte("benchmark-key-25-bytes!!"),
		value:         benchPayload(64),
		subtreeHeight: 0,
		size:          1,
		nodeKey:       NewNodeKey(1, 1),
		useBlake3:     useBlake3,
	}
}

func benchV2Inner(useBlake3 bool) *Node {
	left := benchV2Leaf(useBlake3)
	left._hash()
	right := benchV2Leaf(useBlake3)
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
		useBlake3:     useBlake3,
	}
}

func BenchmarkNode_hash(b *testing.B) {
	for _, blake := range []bool{false, true} {
		algo := "sha256"
		if blake {
			algo = "blake3"
		}
		b.Run(algo+"/leaf", func(b *testing.B) {
			node := benchV2Leaf(blake)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.hash = nil
				hashSink = node._hash()
			}
		})
		b.Run(algo+"/inner", func(b *testing.B) {
			node := benchV2Inner(blake)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				node.hash = nil
				hashSink = node._hash()
			}
		})
	}
}
